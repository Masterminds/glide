package action

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/godep"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
	"github.com/Masterminds/glide/util"
	"github.com/Masterminds/semver"
)

// Get fetches one or more dependencies and installs.
//
// This includes resolving dependency resolution and re-generating the lock file.
func Get(names []string, installer *repo.Installer, insecure, skipRecursive, stripVendor, nonInteract, testDeps bool) {
	cache.SystemLock()

	base := gpath.Basepath()
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()
	glidefile, err := gpath.Glide()
	if err != nil {
		msg.Die("Could not find Glide file: %s", err)
	}

	// Add the packages to the config.
	if count, err2 := addPkgsToConfig(conf, names, insecure, nonInteract, testDeps); err2 != nil {
		msg.Die("Failed to get new packages: %s", err2)
	} else if count == 0 {
		msg.Warn("Nothing to do")
		return
	}

	// Fetch the new packages. Can't resolve versions via installer.Update if
	// get is called while the vendor/ directory is empty so we checkout
	// everything.
	err = installer.Checkout(conf)
	if err != nil {
		msg.Die("Failed to checkout packages: %s", err)
	}

	// Prior to resolving dependencies we need to start working with a clone
	// of the conf because we'll be making real changes to it.
	confcopy := conf.Clone()

	if !skipRecursive {
		// Get all repos and update them.
		// TODO: Can we streamline this in any way? The reason that we update all
		// of the dependencies is that we need to re-negotiate versions. For example,
		// if an existing dependency has the constraint >1.0 and this new package
		// adds the constraint <2.0, then this may re-resolve the existing dependency
		// to be between 1.0 and 2.0. But changing that dependency may then result
		// in that dependency's dependencies changing... so we sorta do the whole
		// thing to be safe.
		err = installer.Update(confcopy, names...)
		if err != nil {
			msg.Die("Could not update packages: %s", err)
		}
	}

	// Set Reference
	if err := repo.SetReference(confcopy, installer.ResolveTest); err != nil {
		msg.Err("Failed to set references: %s", err)
	}

	err = installer.Export(confcopy)
	if err != nil {
		msg.Die("Unable to export dependencies to vendor directory: %s", err)
	}

	// Write YAML
	if err := conf.WriteFile(glidefile); err != nil {
		msg.Die("Failed to write glide YAML file: %s", err)
	}
	if !skipRecursive {
		// Write lock
		if stripVendor {
			confcopy = godep.RemoveGodepSubpackages(confcopy)
		}
		writeLock(conf, confcopy, base)
	} else {
		msg.Warn("Skipping lockfile generation because full dependency tree is not being calculated")
	}

	if stripVendor {
		msg.Info("Removing nested vendor and Godeps/_workspace directories...")
		err := gpath.StripVendor()
		if err != nil {
			msg.Err("Unable to strip vendor directories: %s", err)
		}
	}
}

func writeLock(conf, confcopy *cfg.Config, base string) {
	hash, err := conf.Hash()
	if err != nil {
		msg.Die("Failed to generate config hash. Unable to generate lock file.")
	}
	lock, err := cfg.NewLockfile(confcopy.Imports, confcopy.DevImports, hash)
	if err != nil {
		msg.Die("Failed to generate lock file: %s", err)
	}
	if err := lock.WriteFile(filepath.Join(base, gpath.LockFile)); err != nil {
		msg.Die("Failed to write glide lock file: %s", err)
	}
}

// addPkgsToConfig adds the given packages to the config file.
//
// Along the way it:
// - ensures that this package is not in the ignore list
// - checks to see if this is already in the dependency list.
// - splits version of of package name and adds the version attribute
// - separates repo from packages
// - sets up insecure repo URLs where necessary
// - generates a list of subpackages
func addPkgsToConfig(conf *cfg.Config, names []string, insecure, nonInteract, testDeps bool) (int, error) {

	if len(names) == 1 {
		msg.Info("Preparing to install %d package.", len(names))
	} else {
		msg.Info("Preparing to install %d packages.", len(names))
	}
	numAdded := 0
	for _, name := range names {
		var version string
		parts := strings.Split(name, "#")
		if len(parts) > 1 {
			name = parts[0]
			version = parts[1]
		}

		msg.Info("Attempting to get package %s", name)

		root, subpkg := util.NormalizeName(name)
		if len(root) == 0 {
			return 0, fmt.Errorf("Package name is required for %q.", name)
		}

		if conf.HasDependency(root) {

			var moved bool
			var dep *cfg.Dependency
			// Move from DevImports to Imports
			if !testDeps && !conf.Imports.Has(root) && conf.DevImports.Has(root) {
				dep = conf.DevImports.Get(root)
				conf.Imports = append(conf.Imports, dep)
				conf.DevImports = conf.DevImports.Remove(root)
				moved = true
				numAdded++
				msg.Info("--> Moving %s from testImport to import", root)
			} else if testDeps && conf.Imports.Has(root) {
				msg.Warn("--> Test dependency %s already listed as import", root)
			}

			// Check if the subpackage is present.
			if subpkg != "" {
				if dep == nil {
					dep = conf.Imports.Get(root)
					if dep == nil && testDeps {
						dep = conf.DevImports.Get(root)
					}
				}
				if dep.HasSubpackage(subpkg) {
					if !moved {
						msg.Warn("--> Package %q is already in glide.yaml. Skipping", name)
					}
				} else {
					dep.Subpackages = append(dep.Subpackages, subpkg)
					msg.Info("--> Adding sub-package %s to existing import %s", subpkg, root)
					numAdded++
				}
			} else if !moved {
				msg.Warn("--> Package %q is already in glide.yaml. Skipping", root)
			}
			continue
		}

		if conf.HasIgnore(root) {
			msg.Warn("--> Package %q is set to be ignored in glide.yaml. Skipping", root)
			continue
		}

		dep := &cfg.Dependency{
			Name: root,
		}

		// When retriving from an insecure location set the repo to the
		// insecure location.
		if insecure {
			dep.Repository = "http://" + root
		}

		if version != "" {
			dep.Reference = version
		} else if !nonInteract {
			getWizard(dep)
		}

		if len(subpkg) > 0 {
			dep.Subpackages = []string{subpkg}
		}

		if dep.Reference != "" {
			msg.Info("--> Adding %s to your configuration with the version %s", dep.Name, dep.Reference)
		} else {
			msg.Info("--> Adding %s to your configuration", dep.Name)
		}

		if testDeps {
			conf.DevImports = append(conf.DevImports, dep)
		} else {
			conf.Imports = append(conf.Imports, dep)
		}
		numAdded++
	}
	return numAdded, nil
}

func getWizard(dep *cfg.Dependency) {
	remote := dep.Remote()

	// Lookup dependency info and store in cache.
	msg.Info("--> Gathering release information for %s", dep.Name)
	wizardFindVersions(dep)

	memlatest := cache.MemLatest(remote)
	if memlatest != "" {
		dres := wizardAskLatest(memlatest, dep)
		if dres {
			dep.Reference = memlatest

			sv, err := semver.NewVersion(dep.Reference)
			if err == nil {
				res := wizardAskRange(sv, dep)
				if res == "m" {
					dep.Reference = "^" + sv.String()
				} else if res == "p" {
					dep.Reference = "~" + sv.String()
				}
			}
		}
	}
}
