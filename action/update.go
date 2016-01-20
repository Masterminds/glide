package action

import (
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
)

// Update updates repos and the lock file from the main glide yaml.
func Update(installer *repo.Installer) {
	base := "."
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()

	installer.RootPackage = conf.Name
	installer.Ignore = conf.Ignore

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
		msg.Error("Failed to set references: %s (Skip to cleanup)", err)
	}

	// Flatten
	// I don't think we need flatten anymore. The installer.Update logic should
	// find and install all of the necessary dependencies. And then the version
	// setting logic should correctly set versions.
	//
	// The edge case, which exist today (but innoculously) is where an older
	// version requires dependencies that a newer repo checkout does not
	// have. In that case, it seems that we may get to the point where we'd
	// have to run Update twice.
	msg.Warn("Flatten not implemented.")

	// Vendored cleanup
	// VendoredCleanup. This should ONLY be run if UpdateVendored was specified.
	if installer.UpdateVendored {
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

	// Write lock
	hash, err := conf.Hash()
	if err != nil {
		msg.Die("Failed to generate config hash. Unable to generate lock file.")
	}
	lock := cfg.NewLockfile(confcopy.Imports, hash)
	if err := lock.WriteFile(filepath.Join(base, gpath.LockFile)); err != nil {
		msg.Error("Could not write lock file to %s: %s", base, err)
		return
	}

	/*
		Does(cmd.VendoredSetup, "cfg").
		Using("conf").From("cxt:cfg").
		Using("update").From("cxt:updateVendoredDeps").

		Does(cmd.UpdateImports, "dependencies").
		Using("conf").From("cxt:cfg").
		Using("force").From("cxt:forceUpdate").
		Using("packages").From("cxt:packages").
		Using("home").From("cxt:home").
		Using("cache").From("cxt:useCache").
		Using("cacheGopath").From("cxt:cacheGopath").
		Using("useGopath").From("cxt:useGopath").
		Does(cmd.SetReference, "version").Using("conf").From("cxt:cfg").
		Does(cmd.Flatten, "flattened").Using("conf").From("cxt:cfg").
		Using("packages").From("cxt:packages").
		Using("force").From("cxt:forceUpdate").
		Using("skip").From("cxt:skipFlatten").
		Using("home").From("cxt:home").
		Using("cache").From("cxt:useCache").
		Using("cacheGopath").From("cxt:cacheGopath").
		Using("useGopath").From("cxt:useGopath").
		Does(cmd.VendoredCleanUp, "_").
		Using("conf").From("cxt:flattened").
		Using("update").From("cxt:updateVendoredDeps").
		Does(cmd.WriteYaml, "out").
		Using("conf").From("cxt:cfg").
		Using("filename").From("cxt:toPath").
		Using("toStdout").From("cxt:toStdout").
		Does(cmd.WriteLock, "lock").
		Using("lockfile").From("cxt:Lockfile").
		Using("skip").From("cxt:skipFlatten")
	*/
}
