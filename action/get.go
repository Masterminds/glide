package action

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
	"github.com/Masterminds/glide/util"
	"github.com/Masterminds/semver"
	"github.com/sdboyer/gps"
)

// Get fetches one or more dependencies and installs.
//
// This includes a solver run and re-generating the lock file.
func Get(names []string, installer *repo.Installer, stripVendor, nonInteract bool) {
	base := gpath.Basepath()
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()

	glidefile, err := gpath.Glide()
	if err != nil {
		msg.Die("Could not find Glide file: %s", err)
	}

	vend, err := gpath.Vendor()
	if err != nil {
		msg.Die("Could not find the vendor dir: %s", err)
	}

	params := gps.SolveParameters{
		RootDir:     filepath.Dir(glidefile),
		ImportRoot:  gps.ProjectRoot(conf.Name),
		Manifest:    conf,
		Trace:       true,
		TraceLogger: log.New(os.Stdout, "", 0),
	}

	// We load the lock file early and bail out if there's a problem, because we
	// don't want a get to just update all deps without the user explictly
	// making that choice.
	if gpath.HasLock(base) {
		params.Lock, err = loadLockfile(base, conf)
		if err != nil {
			msg.Err("Could not load lockfile; aborting get. Existing dependency versions cannot be safely preserved without a lock file. Error was: %s", err)
			return
		}
	}

	// Create the SourceManager for this run
	sm, err := gps.NewSourceManager(dependency.Analyzer{}, filepath.Join(installer.Home, "cache"), false)
	defer sm.Release()
	if err != nil {
		msg.Err(err.Error())
		return
	}

	// Now, with the easy/fast errors out of the way, dive into adding the new
	// deps to the manifest.

	// Add the packages to the config.
	//if count, err2 := addPkgsToConfig(conf, names, insecure, nonInteract, testDeps); err2 != nil {
	if count, err2 := addPkgsToConfig(conf, names, false, nonInteract, false); err2 != nil {
		msg.Die("Failed to get new packages: %s", err2)
	} else if count == 0 {
		msg.Warn("Nothing to do")
		return
	}

	// Prepare a solver. This validates our params.
	s, err := gps.Prepare(params, sm)
	if err != nil {
		msg.Err("Aborted get - could not set up solver to reconcile dependencies: %s", err)
		return
	}

	r, err := s.Solve()
	if err != nil {
		// TODO better error handling
		msg.Err("Failed to find a solution for all new dependencies: %s", err.Error())
		return
	}

	// Solve succeeded. Write out the yaml, lock, and vendor to a tmpdir, then mv
	// them all into place iff all the writes worked

	gw := safeGroupWriter{
		conf:        conf,
		lock:        params.Lock.(*cfg.Lockfile),
		resultLock:  r,
		sm:          sm,
		glidefile:   glidefile,
		vendor:      vend,
		stripVendor: stripVendor,
	}

	err = gw.writeAllSafe()
	if err != nil {
		msg.Err(err.Error())
		return
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
	// TODO refactor this to take and use a gps.SourceManager
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

		root, _ := util.NormalizeName(name)
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

			if !moved {
				msg.Warn("--> Package %q is already in glide.yaml. Skipping", root)
			}
			continue
		}

		if conf.HasIgnore(root) {
			msg.Warn("--> Package %q is set to be ignored in glide.yaml. Skipping", root)
			continue
		}

		dep := &cfg.Dependency{
			Name:       root,
			Constraint: gps.Any(),
		}

		// When retriving from an insecure location set the repo to the
		// insecure location.
		if insecure {
			dep.Repository = "http://" + root
		}

		if version != "" {
			// TODO(sdboyer) set the right type...what is that here?
			dep.Constraint = gps.NewVersion(version)
		} else if !nonInteract {
			getWizard(dep)
		}

		if dep.Constraint != nil {
			msg.Info("--> Adding %s to your configuration with the version %s", dep.Name, dep.Constraint)
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
	var remote string
	if dep.Repository != "" {
		remote = dep.Repository
	} else {
		remote = "https://" + dep.Name
	}

	// Lookup dependency info and store in cache.
	msg.Info("--> Gathering release information for %s", dep.Name)
	wizardFindVersions(dep)

	memlatest := cache.MemLatest(remote)
	if memlatest != "" {
		dres := wizardAskLatest(memlatest, dep)
		if dres {
			// TODO(sdboyer) set the right type...what is that here?
			v := gps.NewVersion(memlatest)
			dep.Constraint = v

			if v.Type() == "semver" {
				sv, _ := semver.NewVersion(memlatest)
				res := wizardAskRange(sv, dep)

				if res == "m" {
					// no errors possible here, if init was valid semver version
					dep.Constraint, _ = gps.NewSemverConstraint("^" + v.String())
				} else if res == "p" {
					// no errors possible here, if init was valid semver version
					dep.Constraint, _ = gps.NewSemverConstraint("~" + v.String())
				}
			}
		}
	}
}
