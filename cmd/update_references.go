package cmd

import (
	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/yaml"
)

// UpdateReferences updates the revision numbers on all of the imports.
//
// If a `packages` list is supplied, only the given base packages will
// be updated.
//
// Params:
// 	- conf (*yaml.Config): Configuration
// 	- packages ([]string): A list of packages to update. Default is all packages.
func UpdateReferences(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", &yaml.Config{}).(*yaml.Config)
	plist := p.Get("packages", []string{}).([]string)

	pkgs := list2map(plist)
	restrict := len(pkgs) > 0

	cwd, err := VendorPath(c)
	if err != nil {
		return false, err
	}

	if len(cfg.Imports) == 0 {
		return cfg, nil
	}

	for _, imp := range cfg.Imports {
		if restrict && !pkgs[imp.Name] {
			Debug("===> Skipping %q", imp.Name)
			continue
		}
		commit, err := VcsLastCommit(imp, cwd)
		if err != nil {
			Warn("Could not get commit on %s: %s", imp.Name, err)
		}
		imp.Reference = commit
	}

	return cfg, nil
}

// list2map takes a list of packages names and creates a map of normalized names.
func list2map(in []string) map[string]bool {
	out := make(map[string]bool, len(in))
	for _, v := range in {
		v, _ := NormalizeName(v)
		out[v] = true
	}
	return out
}
