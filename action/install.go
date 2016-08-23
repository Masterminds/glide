package action

import (
	"path/filepath"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
)

// Install installs a vendor directory based on an existing Glide configuration.
func Install(installer *repo.Installer, stripVendor bool) {
	cache.SystemLock()

	base := "."
	// Ensure GOPATH
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()

	// Lockfile exists
	if !gpath.HasLock(base) {
		msg.Info("Lock file (glide.lock) does not exist. Performing update.")
		Update(installer, false, stripVendor)
		return
	}
	// Load lockfile
	lock, err := cfg.ReadLockFile(filepath.Join(base, gpath.LockFile))
	if err != nil {
		msg.Die("Could not load lockfile.")
	}
	// Verify lockfile hasn't changed
	hash, err := conf.Hash()
	if err != nil {
		msg.Die("Could not load lockfile.")
	} else if hash != lock.Hash {
		msg.Warn("Lock file may be out of date. Hash check of YAML failed. You may need to run 'update'")
	}

	// Install
	newConf, err := installer.Install(lock, conf)
	if err != nil {
		msg.Die("Failed to install: %s", err)
	}

	msg.Info("Setting references.")

	// Set reference
	if err := repo.SetReference(newConf, installer.ResolveTest); err != nil {
		msg.Die("Failed to set references: %s (Skip to cleanup)", err)
	}

	err = installer.Export(newConf)
	if err != nil {
		msg.Die("Unable to export dependencies to vendor directory: %s", err)
	}

	if stripVendor {
		msg.Info("Removing nested vendor and Godeps/_workspace directories...")
		err := gpath.StripVendor()
		if err != nil {
			msg.Err("Unable to strip vendor directories: %s", err)
		}
	}
}
