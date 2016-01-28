package action

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
	"github.com/Masterminds/glide/util"
)

// Get fetches one or more dependencies and installs.
//
// This includes resolving dependency resolution and re-generating the lock file.
func Get(names []string, installer *repo.Installer, insecure, skipRecursive bool) {
	base := gpath.Basepath()
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()
	glidefile, err := gpath.Glide()
	if err != nil {
		msg.Die("Could not find Glide file: %s", err)
	}

	// Add the packages to the config.
	if err := addPkgsToConfig(conf, names, insecure); err != nil {
		msg.Die("Failed to get new packages: %s", err)
	}

	// Fetch the new packages. Can't resolve versions via installer.Update if
	// get is called while the vendor/ directory is empty so we checkout
	// everything.
	installer.Checkout(conf, false)

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
		err = installer.Update(confcopy)
		if err != nil {
			msg.Die("Could not update packages: %s", err)
		}
	}

	// Set Reference
	if err := repo.SetReference(confcopy); err != nil {
		msg.Error("Failed to set references: %s", err)
	}

	// VendoredCleanup
	if installer.UpdateVendored {
		repo.VendoredCleanup(confcopy)
	}

	// Write YAML
	if err := conf.WriteFile(glidefile); err != nil {
		msg.Die("Failed to write glide YAML file: %s", err)
	}
	if !skipRecursive {
		// Write lock
		writeLock(conf, confcopy, base)
	} else {
		msg.Warn("Skipping lockfile generation because full dependency tree is not being calculated")
	}
}

func writeLock(conf, confcopy *cfg.Config, base string) {
	hash, err := conf.Hash()
	if err != nil {
		msg.Die("Failed to generate config hash. Unable to generate lock file.")
	}
	lock := cfg.NewLockfile(confcopy.Imports, hash)
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
// - seperates repo from packages
// - sets up insecure repo URLs where necessary
// - generates a list of subpackages
func addPkgsToConfig(conf *cfg.Config, names []string, insecure bool) error {

	msg.Info("Preparing to install %d package.", len(names))

	for _, name := range names {
		var version string
		parts := strings.Split(name, "#")
		if len(parts) > 1 {
			name = parts[0]
			version = parts[1]
		}

		root := util.GetRootFromPackage(name)
		if len(root) == 0 {
			return fmt.Errorf("Package name is required for %q.", name)
		}

		if conf.HasDependency(root) {

			// Check if the subpackage is present.
			subpkg := strings.TrimPrefix(name, root)
			subpkg = strings.TrimPrefix(subpkg, "/")
			if subpkg != "" {
				found := false
				dep := conf.Imports.Get(root)
				for _, s := range dep.Subpackages {
					if s == subpkg {
						found = true
						break
					}
				}
				if found {
					msg.Warn("Package %q is already in glide.yaml. Skipping", name)
				} else {
					dep.Subpackages = append(dep.Subpackages, subpkg)
					msg.Info("Adding sub-package %s to existing import %s", subpkg, root)
				}
			} else {
				msg.Warn("Package %q is already in glide.yaml. Skipping", root)
			}
			continue
		}

		if conf.HasIgnore(root) {
			msg.Warn("Package %q is set to be ignored in glide.yaml. Skipping", root)
			continue
		}

		dep := &cfg.Dependency{
			Name: root,
		}

		if version != "" {
			dep.Reference = version
		}

		// When retriving from an insecure location set the repo to the
		// insecure location.
		if insecure {
			dep.Repository = "http://" + root
		}

		subpkg := strings.TrimPrefix(name, root)
		if len(subpkg) > 0 && subpkg != "/" {
			dep.Subpackages = []string{strings.TrimPrefix(subpkg, "/")}
		}

		if dep.Reference != "" {
			msg.Info("Importing %s with the version %s", dep.Name, dep.Reference)
		} else {
			msg.Info("Importing %s", dep.Name)
		}

		conf.Imports = append(conf.Imports, dep)
	}
	return nil
}
