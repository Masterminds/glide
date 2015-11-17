package cmd

import (
	"path"

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
	vend, _ := VendorPath(c)
	pkgs := list2map(plist)
	restrict := len(pkgs) > 0

	cwd, err := VendorPath(c)
	if err != nil {
		return false, err
	}

	if len(cfg.Imports) == 0 {
		return cfg, nil
	}

	// Walk the dependency tree to discover all the packages to pin.
	packages := make([]string, len(cfg.Imports))
	for i, v := range cfg.Imports {
		packages[i] = v.Name
	}
	deps := make(map[string]*yaml.Dependency, len(cfg.Imports))
	for _, imp := range cfg.Imports {
		deps[imp.Name] = imp
	}
	f := &flattening{cfg, vend, vend, deps, packages}
	err = discoverDependencyTree(f)
	if err != nil {
		return cfg, err
	}

	exportFlattenedDeps(cfg, deps)

	err = cfg.DeDupe()
	if err != nil {
		return cfg, err
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

func discoverDependencyTree(f *flattening) error {
	Debug("---> Inspecting %s for dependencies (%d packages).\n", f.curr, len(f.scan))
	for _, imp := range f.scan {
		Debug("----> Scanning %s", imp)
		base := path.Join(f.top, imp)
		mod := []string{}
		if m, ok := mergeGlide(base, imp, f.deps, f.top); ok {
			mod = m
		} else if m, ok = mergeGodep(base, imp, f.deps, f.top); ok {
			mod = m
		} else if m, ok = mergeGPM(base, imp, f.deps, f.top); ok {
			mod = m
		} else if m, ok = mergeGb(base, imp, f.deps, f.top); ok {
			mod = m
		} else if m, ok = mergeGuess(base, imp, f.deps, f.top); ok {
			mod = m
		}

		if len(mod) > 0 {
			Debug("----> Looking for dependencies in %q (%d)", imp, len(mod))
			f2 := &flattening{
				conf: f.conf,
				top:  f.top,
				curr: base,
				deps: f.deps,
				scan: mod}
			discoverDependencyTree(f2)
		}
	}

	return nil
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
