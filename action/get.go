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
func Get(names []string, installer *repo.Installer, insecure bool) {
	base := gpath.Basepath()
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()
	glidefile, err := gpath.Glide()
	if err != nil {
		msg.Die("Could not find Glide file: %s", err)
	}

	// Add the packages to the config.
	deps, err := addPkgsToConfig(conf, names, insecure)
	if err != nil {
		msg.Die("Failed to get new packages: %s", err)
	}
	conf.Imports = deps

	// Get all repos and update them.
	// TODO: Can we streamline this in any way? The reason that we update all
	// of the dependencies is that we need to re-negotiate versions. For example,
	// if an existing dependency has the constraint >1.0 and this new package
	// adds the constraint <2.0, then this may re-resolve the existing dependency
	// to be betwee 1.0 and 2.0. But changing that dependency may then result
	// in that dependency's dependencies changing... so we sorta do the whole
	// thing to be safe.
	lock, err := installer.Update(conf)
	if err != nil {
		msg.Die("Could not update packages: %s", err)
	}

	// Set Reference
	if err := repo.SetReference(conf); err != nil {
		msg.Error("Failed to set references: %s", err)
	}

	// Flatten
	// Flatten is not implemented right now because I think Update handles it.

	// VendoredCleanup
	if installer.UpdateVendored {
		repo.VendoredCleanup(conf)
	}

	// Write YAML
	if err := conf.WriteFile(glidefile); err != nil {
		msg.Die("Failed to write glide YAML file: %s", err)
	}

	// Write lock
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
func addPkgsToConfig(conf *cfg.Config, names []string, insecure bool) ([]*cfg.Dependency, error) {

	msg.Info("Preparing to install %d package.", len(names))

	deps := []*cfg.Dependency{}
	for _, name := range names {
		var version string
		parts := strings.Split(name, "#")
		if len(parts) > 1 {
			name = parts[0]
			version = parts[1]
		}

		root := util.GetRootFromPackage(name)
		if len(root) == 0 {
			return nil, fmt.Errorf("Package name is required for %q.", name)
		}

		if conf.HasDependency(root) {
			msg.Warn("Package %q is already in glide.yaml. Skipping", root)
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
			dep.Subpackages = []string{subpkg}
		}

		if dep.Reference != "" {
			msg.Info("Importing %s with the version %s", dep.Name, dep.Reference)
		} else {
			msg.Info("Importing %s", dep.Name)
		}

		// FIXME: I don't think we need to do this anymore. I think we only
		// need to manage the `deps` list.
		conf.Imports = append(conf.Imports, dep)
		deps = append(deps, dep)
	}
	return deps, nil
}
