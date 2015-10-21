package cmd

import (
	"os"
	"path"

	"github.com/Masterminds/cookoo"
	"github.com/kylelemons/go-gypsy/yaml"
)

// Flatten recurses through all dependent packages and flattens to a top level.
//
// Flattening involves determining a tree's dependencies and flattening them
// into a single large list.
//
// Params:
//	- packages ([]string): The packages to read. If this is empty, it reads all
//		packages.
//	- force (bool): force git updates.
//	- conf (*Config): The configuration.
//
// Returns:
//
func Flatten(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	conf := p.Get("conf", &Config{}).(*Config)
	skip := p.Get("skip", false).(bool)
	if skip {
		return conf, nil
	}
	packages := p.Get("packages", []string{}).([]string)
	force := p.Get("force", true).(bool)
	vend, _ := VendorPath(c)

	// If no packages are supplied, we do them all.
	if len(packages) == 0 {
		packages = make([]string, len(conf.Imports))
		for i, v := range conf.Imports {
			packages[i] = v.Name
		}
	}

	// Build an initial dependency map.
	deps := make(map[string]*Dependency, len(conf.Imports))
	for _, imp := range conf.Imports {
		deps[imp.Name] = imp
	}

	f := &flattening{conf, vend, vend, deps, packages}

	err := recFlatten(f, force)
	flattenSetRefs(f)
	Info("Project relies on %d dependencies.", len(deps))
	exportFlattenedDeps(conf, deps)

	return conf, err
}

func exportFlattenedDeps(conf *Config, in map[string]*Dependency) {
	out := make([]*Dependency, len(in))
	i := 0
	for _, v := range in {
		out[i] = v
		i++
	}
	conf.Imports = out
}

type flattening struct {
	conf *Config
	// Top vendor path, e.g. project/vendor
	top string
	// Current path
	curr string
	// Built list of dependencies
	deps map[string]*Dependency
	// Dependencies that need to be scanned.
	scan []string
}

// Hack: Cache record of updates so we don't have to keep doing git pulls.
var updateCache = map[string]bool{}

// refFlatten recursively flattens the vendor tree.
func recFlatten(f *flattening, force bool) error {
	Debug("---> Inspecting %s for changes (%d packages).\n", f.curr, len(f.scan))
	for _, imp := range f.scan {
		Debug("----> Scanning %s", imp)
		base := path.Join(f.top, imp)
		mod := []string{}
		if m, ok := mergeGlide(base, imp, f.deps); ok {
			mod = m
		} else if m, ok = mergeGodep(base, imp, f.deps); ok {
			mod = m
		} else if m, ok = mergeGPM(base, imp, f.deps); ok {
			mod = m
		} else if m, ok = mergeGb(base, imp, f.deps); ok {
			mod = m
		} else if m, ok = mergeGuess(base, imp, f.deps); ok {
			mod = m
		}

		if len(mod) > 0 {
			Debug("----> Updating all dependencies for %q (%d)", imp, len(mod))
			flattenGlideUp(f, base, force)
			f2 := &flattening{
				conf: f.conf,
				top:  f.top,
				curr: base,
				deps: f.deps,
				scan: mod}
			recFlatten(f2, force)
		}
	}

	return nil
}

// flattenGlideUp does a glide update in the middle of a flatten operation.
//
// While this is expensive, it is also necessary to make sure we have the
// correct version of all dependencies. We might be able to simplify by
// marking packages dirty when they are added.
func flattenGlideUp(f *flattening, base string, force bool) error {
	//vdir := path.Join(base, "vendor")
	for _, imp := range f.deps {
		wd := path.Join(f.top, imp.Name)
		if VcsExists(imp, wd) {
			if updateCache[imp.Name] {
				Debug("----> Already updated %s", imp.Name)
				continue
			}
			Debug("Updating project %s (%s)\n", imp.Name, wd)
			if err := VcsUpdate(imp, f.top, force); err != nil {
				// We can still go on just fine even if this fails.
				Warn("Skipped update %s: %s\n", imp.Name, err)
				continue
			}
			updateCache[imp.Name] = true
		} else {
			Debug("Importing %s to project %s\n", imp.Name, wd)
			if err := VcsGet(imp, wd); err != nil {
				Warn("Skipped getting %s: %v\n", imp.Name, err)
				continue
			}
		}

		// If a revision has been set use it.
		err := VcsVersion(imp, f.top)
		if err != nil {
			Warn("Problem setting version on %s: %s\n", imp.Name, err)
		}
	}

	return nil
}

// Set the references for all packages after a flatten is completed.
func flattenSetRefs(f *flattening) {
	Debug("Setting final version for %d dependencies.", len(f.deps))
	for _, imp := range f.deps {
		if err := VcsVersion(imp, f.top); err != nil {
			Warn("Problem setting version on %s: %s (flatten)\n", imp.Name, err)
		}
	}
}

func mergeGlide(dir, name string, deps map[string]*Dependency) ([]string, bool) {
	gp := path.Join(dir, "glide.yaml")
	if _, err := os.Stat(gp); err != nil {
		return []string{}, false
	}
	f, err := yaml.ReadFile(gp)
	if err != nil {
		Warn("Found glide file %q, but can't parse: %s", gp, err)
		return []string{}, false
	}

	conf, err := FromYaml(f.Root)
	if err != nil {
		Warn("Found glide file %q, but can't use it: %s", gp, err)
		return []string{}, false
	}

	Info("Found glide.yaml in %s", gp)

	return mergeDeps(deps, conf.Imports), true
}

// listGodep appends Godeps entries to the deps.
//
// It returns true if any dependencies were found (even if not added because
// they are duplicates).
func mergeGodep(dir, name string, deps map[string]*Dependency) ([]string, bool) {
	Debug("Looking in %s/Godeps/ for a Godeps.json file.\n", dir)
	d, err := parseGodepGodeps(dir)
	if err != nil {
		Warn("Looking for Godeps: %s\n", err)
		return []string{}, false
	} else if len(d) == 0 {
		return []string{}, false
	}

	Info("Found Godeps.json file for %q", name)
	return mergeDeps(deps, d), true
}

// listGb merges GB dependencies into the deps.
func mergeGb(dir, pkg string, deps map[string]*Dependency) ([]string, bool) {
	Debug("Looking in %s/vendor/ for a manifest file.\n", dir)
	d, err := parseGbManifest(dir)
	if err != nil || len(d) == 0 {
		return []string{}, false
	}
	Info("Found gb manifest file for %q", pkg)
	return mergeDeps(deps, d), true
}

// mergeGPM merges GPM Godeps files into deps.
func mergeGPM(dir, pkg string, deps map[string]*Dependency) ([]string, bool) {
	d, err := parseGPMGodeps(dir)
	if err != nil || len(d) == 0 {
		return []string{}, false
	}
	Info("Found GPM file for %q", pkg)
	return mergeDeps(deps, d), true
}

// mergeGuess guesses dependencies and merges.
//
// This always returns true because it always handles the job of searching
// for dependencies. So generally it should be the last merge strategy
// that you try.
func mergeGuess(dir, pkg string, deps map[string]*Dependency) ([]string, bool) {
	/*
			Info("Scanning %s for dependencies.", pkg)
			buildContext, err := GetBuildContext()
			if err != nil {
				Warn("Could not scan package %q: %s", pkg, err)
				return []string{}, false
			}

			res := []string{}

			if _, err := os.Stat(dir); err != nil {
				Warn("Directory is missing: %s", dir)
				return res, true
			}

			d := walkDeps(buildContext, dir, pkg)
			for _, name := range d {
				name, _ := NormalizeName(name)
				repo := getRepoRootFromPackage(name)
				if _, ok := deps[name]; ok {
					Debug("====> Seen %s already. Skipping", name)
					continue
				}

				found := findPkg(buildContext, name, dir)
				switch found.PType {
				case ptypeUnknown:
					Debug("✨☆ Undownloaded dependency: %s", name)
					nd := &Dependency{
						Name:       name,
						Repository: "https://" + repo,
					}
					deps[name] = nd
					res = append(res, name)
				case ptypeGoroot, ptypeCgo:
					break
				default:
					// We're looking for dependencies that might exist in $GOPATH
					// but not be on vendor. We add any that are on $GOPATH.
					if _, ok := deps[name]; !ok {
						Debug("✨☆ GOPATH dependency: %s", name)
						nd := &Dependency{Name: name}
						deps[name] = nd
						res = append(res, name)
					}
				}
			}

		return res, true
	*/
	Info("Package %s manages its own dependencies", pkg)
	return []string{}, true
}

// mergeDeps merges any dependency array into deps.
func mergeDeps(orig map[string]*Dependency, add []*Dependency) []string {
	mod := []string{}
	for _, dd := range add {
		// Add it unless it's already there.
		if existing, ok := orig[dd.Name]; !ok {
			orig[dd.Name] = dd
			Debug("Adding %s to the scan list", dd.Name)
			mod = append(mod, dd.Name)
		} else if existing.Reference == "" && dd.Reference != "" {
			// If a nested dep has finer dependency references than outside,
			// set the reference.
			existing.Reference = dd.Reference
			mod = append(mod, dd.Name)
		} else if dd.Reference != "" && existing.Reference != "" && dd.Reference != existing.Reference {
			// We can detect version conflicts, but we can't really do
			// anything to correct, since we don't know the intentions of the
			// authors.
			Warn("Conflict: %s ref is %s, but also asked for %s", existing.Name, existing.Reference, dd.Reference)
			Info("Keeping %s %s", existing.Name, existing.Reference)
		}
	}
	return mod
}
