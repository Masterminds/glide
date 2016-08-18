package action

import (
	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
)

// Remove removes a dependncy from the configuration.
func Remove(packages []string, inst *repo.Installer) {
	cache.SystemLock()
	base := gpath.Basepath()
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()
	glidefile, err := gpath.Glide()
	if err != nil {
		msg.Die("Could not find Glide file: %s", err)
	}

	msg.Info("Preparing to remove %d packages.", len(packages))
	conf.Imports = rmDeps(packages, conf.Imports)
	conf.DevImports = rmDeps(packages, conf.DevImports)

	// Copy used to generate locks.
	confcopy := conf.Clone()

	//confcopy.Imports = inst.List(confcopy)

	if err := repo.SetReference(confcopy, inst.ResolveTest); err != nil {
		msg.Err("Failed to set references: %s", err)
	}

	err = inst.Export(confcopy)
	if err != nil {
		msg.Die("Unable to export dependencies to vendor directory: %s", err)
	}

	// Write glide.yaml
	if err := conf.WriteFile(glidefile); err != nil {
		msg.Die("Failed to write glide YAML file: %s", err)
	}

	// Write glide lock
	writeLock(conf, confcopy, base)
}

// rmDeps returns a list of dependencies that do not contain the given pkgs.
//
// It generates neither an error nor a warning for a pkg that does not exist
// in the list of deps.
func rmDeps(pkgs []string, deps []*cfg.Dependency) []*cfg.Dependency {
	res := []*cfg.Dependency{}
	for _, d := range deps {
		rem := false
		for _, p := range pkgs {
			if p == d.Name {
				rem = true
			}
		}
		if !rem {
			res = append(res, d)
		}
	}
	return res
}
