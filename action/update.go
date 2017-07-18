package action

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
)

var mainPkgRegex = regexp.MustCompile(`(?m)^package\s*main\s*$`)

// Update updates repos and the lock file from the main glide yaml.
func Update(installer *repo.Installer, skipRecursive, stripVendor bool) {
	cache.SystemLock()

	base := "."
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()

	// Try to check out the initial dependencies.
	if err := installer.Checkout(conf); err != nil {
		msg.Die("Failed to do initial checkout of config: %s", err)
	}

	// Set the versions for the initial dependencies so that resolved dependencies
	// are rooted in the correct version of the base.
	if err := repo.SetReference(conf, installer.ResolveTest); err != nil {
		msg.Die("Failed to set initial config references: %s", err)
	}

	// Prior to resolving dependencies we need to start working with a clone
	// of the conf because we'll be making real changes to it.
	confcopy := conf.Clone()

	if !skipRecursive {
		// Get all repos and update them.
		err := installer.Update(confcopy)
		if err != nil {
			msg.Die("Could not update packages: %s", err)
		}

		// Set references. There may be no remaining references to set since the
		// installer set them as it went to make sure it parsed the right imports
		// from the right version of the package.
		msg.Info("Setting references for remaining imports")
		if err := repo.SetReference(confcopy, installer.ResolveTest); err != nil {
			msg.Err("Failed to set references: %s (Skip to cleanup)", err)
		}
	}

	err := installer.Export(confcopy)
	if err != nil {
		msg.Die("Unable to export dependencies to vendor directory: %s", err)
	}

	// Write glide.yaml (Why? Godeps/GPM/GB?)
	// I think we don't need to write a new Glide file because update should not
	// change anything important. It will just generate information about
	// transative dependencies, all of which belongs exclusively in the lock
	// file, not the glide.yaml file.
	// TODO(mattfarina): Detect when a new dependency has been added or removed
	// from the project. A removed dependency should warn and an added dependency
	// should be added to the glide.yaml file. See issue #193.

	if !skipRecursive {
		// Write lock
		hash, err := conf.Hash()
		if err != nil {
			msg.Die("Failed to generate config hash. Unable to generate lock file.")
		}
		lock, err := cfg.NewLockfile(confcopy.Imports, confcopy.DevImports, hash)
		if err != nil {
			msg.Die("Failed to generate lock file: %s", err)
		}
		wl := true
		if gpath.HasLock(base) {
			yml, err := ioutil.ReadFile(filepath.Join(base, gpath.LockFile))
			if err == nil {
				l2, err := cfg.LockfileFromYaml(yml)
				if err == nil {
					f1, err := l2.Fingerprint()
					f2, err2 := lock.Fingerprint()
					if err == nil && err2 == nil && f1 == f2 {
						wl = false
					}
				}
			}
		}
		if wl {
			if err := lock.WriteFile(filepath.Join(base, gpath.LockFile)); err != nil {
				msg.Err("Could not write lock file to %s: %s", base, err)
				return
			}
		} else {
			msg.Info("Versions did not change. Skipping glide.lock update.")
		}

		msg.Info("Project relies on %d dependencies.", len(confcopy.Imports))
	} else {
		msg.Warn("Skipping lockfile generation because full dependency tree is not being calculated")
	}

	if stripVendor {
		msg.Info("Removing nested vendor and Godeps/_workspace directories...")
		err := gpath.StripVendor()
		if err != nil {
			msg.Err("Unable to strip vendor directories: %s", err)
		}
		msg.Info("Cleaning test and unnecessary files from vendor directories...")
		err = CleanVendor(confcopy)
		if err != nil {
			msg.Err("Unable to clean vendor directories: %s", err)
		}
	}
}

func CleanVendor(conf *cfg.Config) error {
	searchPath, _ := gpath.Vendor()
	if _, err := os.Stat(searchPath); err != nil {
		if os.IsNotExist(err) {
			msg.Debug("Vendor directory does not exist.")
		}
		return err
	}

	baseDir, err := filepath.Abs(".")
	if err != nil {
		msg.Die("Could not read directory: %s", err)
	}

	res, err := dependency.NewResolver(baseDir)
	if err != nil {
		msg.Die("Failed to create a resolver: %s", err)
	}
	res.Handler = &dependency.DefaultMissingPackageHandler{
		Prefix:  path.Join(baseDir, "vendor"),
		Missing: []string{},
		Gopath:  []string{},
	}
	res.ResolveTest = true
	res.Config = conf
	msg.Info("Resolving vendored imports")

	imports, testImports, err := res.ResolveLocal(true)
	if err != nil {
		msg.Die("Failed to resolve vendored packages: %s", err)
	}

	imports = append(imports, testImports...)

	fmt.Printf("Scanned Imports: %+v\n", imports)

	return filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		// Skip the base vendor directory
		if path == searchPath {
			return nil
		}

		// Skip paths we have already deleted
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil
		}

		if info.IsDir() {
			pkg := path[len(searchPath)+1:]
			if !isDirDependency(imports, pkg) {
				msg.Debug("Removing pkg: %s", pkg)
				return os.RemoveAll(path)
			}
		} else {

			normalizedPath := strings.ToLower(path)
			if strings.HasSuffix(normalizedPath, "_test.go") {
				msg.Debug("Removing Test: %s", path)
				return os.RemoveAll(path)
			}
			// TODO: Provide the user an option to keep license and legal notices around
			if isSrcFile(normalizedPath) {
				if isMainPkg(path) {
					msg.Debug("Removing main package: %s", path)
					return os.Remove(path)
				}
				// If the files dir is not in our dependency list
				pkg := filepath.Dir(path[len(searchPath)+1:])
				msg.Debug("Src PKG: %s - %s", pkg, path)
				if !isSrcDependency(imports, pkg) {
					msg.Debug("Removing NonDepencency Src: %s", path)
					os.Remove(path)
				}
				return nil
			}
			msg.Debug("Removing file: %s", path)
			return os.RemoveAll(path)
		}
		return nil
	})
}

// Return true if the directory provided matches or is part of a path to any
// of the packages listed in our dependencies
func isDirDependency(deps []string, dir string) bool {
	for _, dep := range deps {
		// If the directory is part of a package path or matches a package exactly
		if strings.HasPrefix(dep, dir) {
			return true
		}
	}
	return false
}

// Return true if the source directory provided is directly specified as a dependency
func isSrcDependency(deps []string, dir string) bool {
	for _, dep := range deps {
		if dep == dir {
			return true
		}
	}
	return false
}

// Return true if the file provided is a source file
func isSrcFile(path string) bool {
	for _, suffix := range []string{".go", ".s"} {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

// Return true if the file provided is a source file
func isMainPkg(path string) bool {
	fd, err := os.Open(path)
	if err != nil {
		msg.Debug("isMainPkg: %s", err)
		return false
	}
	contents, err := ioutil.ReadAll(fd)
	if err != nil {
		msg.Debug("isMainPkg: %s", err)
		return false
	}

	if mainPkgRegex.Match(contents) {
		return true
	}
	return false
}
