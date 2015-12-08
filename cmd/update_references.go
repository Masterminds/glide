package cmd

import (
	"path/filepath"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
)

// UpdateReferences updates the revision numbers on all of the imports.
//
// If a `packages` list is supplied, only the given base packages will
// be updated.
//
// Params:
// 	- conf (*cfg.Config): Configuration
// 	- packages ([]string): A list of packages to update. Default is all packages.
func UpdateReferences(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	conf := p.Get("conf", &cfg.Config{}).(*cfg.Config)
	plist := p.Get("packages", []string{}).([]string)
	vend, _ := VendorPath(c)
	pkgs := list2map(plist)
	restrict := len(pkgs) > 0

	cwd, err := VendorPath(c)
	if err != nil {
		return false, err
	}

	if len(conf.Imports) == 0 {
		return conf, nil
	}

	// Walk the dependency tree to discover all the packages to pin.
	packages := make([]string, len(conf.Imports))
	for i, v := range conf.Imports {
		packages[i] = v.Name
	}
	deps := make(map[string]*cfg.Dependency, len(conf.Imports))
	for _, imp := range conf.Imports {
		deps[imp.Name] = imp
	}
	f := &flattening{conf, vend, vend, deps, packages}
	err = discoverDependencyTree(f)
	if err != nil {
		return conf, err
	}

	exportFlattenedDeps(conf, deps)

	err = conf.DeDupe()
	if err != nil {
		return conf, err
	}

	for _, imp := range conf.Imports {
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

	return conf, nil
}

func discoverDependencyTree(f *flattening) error {
	Debug("---> Inspecting %s for dependencies (%d packages).\n", f.curr, len(f.scan))

	// projects tracks which projects we have seen. Initially, it's everytning
	// in f.deps. As we go, we'll merge in all of the others that we find.
	projects := f.deps

	// Get all of the packages that are used.
	resolver, err := dependency.NewResolver(f.top)
	if err != nil {
		return err
	}

	dlist := make([]*cfg.Dependency, 0, len(f.deps))
	for _, d := range f.deps {
		dlist = append(dlist, d)
	}
	pkgs, err := resolver.ResolveAll(dlist)
	if err != nil {
		return err
	}

	// From the packages, we just want the repositories. So we get a normalized
	// list of dependencies.
	for _, d := range pkgs {
		d, err = filepath.Rel(f.top, d)
		if err != nil {
			Warn("Cannot resolve relative path: %s", err)
		}
		d, _ := NormalizeName(d)

		if _, ok := projects[d]; !ok {
			projects[d] = &cfg.Dependency{Name: d}
			Info("====> %s", d)
		}
	}

	// At this point, we know that we have an exhaustive list of packages, so
	// we can now just look for files that will tell us what version of each
	// package to use.

	/*
		scanned := map[string]bool{}
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
			} else if m, ok = mergeGuess(base, imp, f.deps, f.top, scanned); ok {
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
	*/

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
