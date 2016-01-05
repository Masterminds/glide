package action

import (
	"io/ioutil"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
)

// Install installs a vendor directory based on an existing Glide configuration.
func Install(installer *repo.Installer) {
	base := "."
	// Ensure GOPATH
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()

	// Lockfile exists
	if !gpath.HasLock(base) {
		msg.Info("Lock file (glide.lock) does not exist. Performing update.")
		msg.Die("No update performed. FIXME.")
		// Update(installer)
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
		cfg.DeleteUnusedPackages(conf)
	}

	// Install
	newConf, err := installer.Install(lock, conf)
	if err != nil {
		msg.Die("Failed to install: %s", err)
	}

	msg.Info("Setting references.")

	// Set reference
	if err := repo.SetReference(newConf); err != nil {
		msg.Error("Failed to set references: %s (Skip to cleanup)", err)
	}

	// VendoredCleanup. This should ONLY be run if UpdateVendored was specified.
	if installer.UpdateVendored {
		repo.VendoredCleanup(newConf)
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
