package action

import (
	"io/ioutil"
	"path/filepath"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
)

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
	}
}
