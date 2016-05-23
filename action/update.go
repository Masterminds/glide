package action

import (
	"io/ioutil"
	"path/filepath"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/godep"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
)

// Update updates repos and the lock file from the main glide yaml.
func Update(installer *repo.Installer, skipRecursive, strip, stripVendor bool) {
	if installer.UseCache {
		cache.SystemLock()
	}

	base := "."
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()

	// Delete unused packages
	if installer.DeleteUnused {
		dependency.DeleteUnused(conf)
	}

	// Try to check out the initial dependencies.
	if err := installer.Checkout(conf, false); err != nil {
		msg.Die("Failed to do initial checkout of config: %s", err)
	}

	// Set the versions for the initial dependencies so that resolved dependencies
	// are rooted in the correct version of the base.
	if err := repo.SetReference(conf); err != nil {
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

		// TODO: There is no support here for importing Godeps, GPM, and GB files.
		// I think that all we really need to do now is hunt for these files, and then
		// roll their version numbers into the config file.

		// Set references. There may be no remaining references to set since the
		// installer set them as it went to make sure it parsed the right imports
		// from the right version of the package.
		msg.Info("Setting references for remaining imports")
		if err := repo.SetReference(confcopy); err != nil {
			msg.Err("Failed to set references: %s (Skip to cleanup)", err)
		}
	}
	// Vendored cleanup
	// VendoredCleanup. This should ONLY be run if UpdateVendored was specified.
	// When stripping VCS happens this will happen as well. No need for double
	// effort.
	if installer.UpdateVendored && !strip {
		repo.VendoredCleanup(confcopy)
	}

	// Write glide.yaml (Why? Godeps/GPM/GB?)
	// I think we don't need to write a new Glide file because update should not
	// change anything important. It will just generate information about
	// transative dependencies, all of which belongs exclusively in the lock
	// file, not the glide.yaml file.
	// TODO(mattfarina): Detect when a new dependency has been added or removed
	// from the project. A removed dependency should warn and an added dependency
	// should be added to the glide.yaml file. See issue #193.

	if stripVendor {
		confcopy = godep.RemoveGodepSubpackages(confcopy)
	}

	if !skipRecursive {
		// Write lock
		hash, err := conf.Hash()
		if err != nil {
			msg.Die("Failed to generate config hash. Unable to generate lock file.")
		}
		lock := cfg.NewLockfile(confcopy.Imports, hash)
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

	if strip {
		msg.Info("Removing version control data from vendor directory...")
		gpath.StripVcs()
	}

	if stripVendor {
		msg.Info("Removing nested vendor and Godeps/_workspace directories...")
		err := gpath.StripVendor()
		if err != nil {
			msg.Err("Unable to strip vendor directories: %s", err)
		}
	}
}
