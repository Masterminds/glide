package action

import (
	"io/ioutil"
	"path/filepath"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
)

// Install installs a vendor directory based on an existing Glide configuration.
func Install(installer *repo.Installer, strip, stripVendor bool) {
	if installer.UseCache {
		cache.SystemLock()
	}

	base := "."
	// Ensure GOPATH
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()

	// Lockfile exists
	if !gpath.HasLock(base) {
		msg.Info("Lock file (glide.lock) does not exist. Performing update.")
		Update(installer, false, strip, stripVendor)
		return
	}
	// Load lockfile
	lock, err := LoadLockfile(base, conf)
	if err != nil {
		msg.Die("Could not load lockfile.")
	}

	// Delete unused packages
	if installer.DeleteUnused {
		// It's unclear whether this should operate off of the lock, or off
		// of the glide.yaml file. I'd think that doing this based on the
		// lock would be much more reliable.
		dependency.DeleteUnused(conf)
	}

	// Install
	newConf, err := installer.Install(lock, conf)
	if err != nil {
		msg.Die("Failed to install: %s", err)
	}

	msg.Info("Setting references.")

	// Set reference
	if err := repo.SetReference(newConf, installer.ResolveTest); err != nil {
		msg.Err("Failed to set references: %s (Skip to cleanup)", err)
	}

	// VendoredCleanup. This should ONLY be run if UpdateVendored was specified.
	// When stripping VCS happens this will happen as well. No need for double
	// effort.
	if installer.UpdateVendored && !strip {
		repo.VendoredCleanup(newConf)
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

// LoadLockfile loads the contents of a glide.lock file.
//
// TODO: This should go in another package.
func LoadLockfile(base string, conf *cfg.Config) (*cfg.Lockfile, error) {
	yml, err := ioutil.ReadFile(filepath.Join(base, gpath.LockFile))
	if err != nil {
		return nil, err
	}
	lock, err := cfg.LockfileFromYaml(yml)
	if err != nil {
		return nil, err
	}

	hash, err := conf.Hash()
	if err != nil {
		return nil, err
	}

	if hash != lock.Hash {
		msg.Warn("Lock file may be out of date. Hash check of YAML failed. You may need to run 'update'")
	}

	return lock, nil
}
