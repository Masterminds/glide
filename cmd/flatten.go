package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/util"
	"github.com/Masterminds/semver"
)

// Flatten recurses through all dependent packages and flattens to a top level.
//
// Flattening involves determining a tree's dependencies and flattening them
// into a single large list.
//
// Params:
//	- packages ([]string): The packages to read. If this is empty, it reads all
//		packages.
//	- force (bool): force vcs updates.
//	- conf (*cfg.Config): The configuration.
//
// Returns:
//
func Flatten(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	conf := p.Get("conf", &cfg.Config{}).(*cfg.Config)
	skip := p.Get("skip", false).(bool)
	home := p.Get("home", "").(string)
	cache := p.Get("cache", false).(bool)
	cacheGopath := p.Get("cacheGopath", false).(bool)
	useGopath := p.Get("useGopath", false).(bool)

	if skip {
		Warn("Skipping lockfile generation because full dependency tree is not being calculated")
		return conf, nil
	}
	packages := p.Get("packages", []string{}).([]string)

	// Operate on a clone of the conf so any changes don't impact later operations.
	// This is a deep clone so dependencies are also cloned.
	confcopy := conf.Clone()

	// Generate a hash of the conf for later use in lockfile generation.
	hash, err := conf.Hash()
	if err != nil {
		return conf, err
	}

	// When packages are passed around with a #version on the end it needs
	// to be stripped.
	for k, v := range packages {
		parts := strings.Split(v, "#")
		packages[k] = parts[0]
	}

	force := p.Get("force", true).(bool)
	vend, _ := VendorPath(c)

	// If no packages are supplied, we do them all.
	if len(packages) == 0 {
		packages = make([]string, len(confcopy.Imports))
		for i, v := range confcopy.Imports {
			packages[i] = v.Name
		}
	}

	// Build an initial dependency map.
	deps := make(map[string]*cfg.Dependency, len(confcopy.Imports))
	for _, imp := range confcopy.Imports {
		deps[imp.Name] = imp
	}

	f := &flattening{confcopy, vend, vend, deps, packages}

	// The assumption here is that once something has been scanned once in a
	// run, there is no need to scan it again.
	scanned := map[string]bool{}
	err = recFlatten(f, force, home, cache, cacheGopath, useGopath, scanned)
	if err != nil {
		return confcopy, err
	}
	err = confcopy.DeDupe()
	if err != nil {
		return confcopy, err
	}
	flattenSetRefs(f)
	Info("Project relies on %d dependencies.", len(deps))

	c.Put("Lockfile", cfg.LockfileFromMap(deps, hash))

	exportFlattenedDeps(confcopy, deps)

	return confcopy, err
}

func exportFlattenedDeps(conf *cfg.Config, in map[string]*cfg.Dependency) {
	out := make([]*cfg.Dependency, len(in))
	i := 0
	for _, v := range in {
		out[i] = v
		i++
	}
	conf.Imports = out
}

type flattening struct {
	conf *cfg.Config
	// Top vendor path, e.g. project/vendor
	top string
	// Current path
	curr string
	// Built list of dependencies
	deps map[string]*cfg.Dependency
	// Dependencies that need to be scanned.
	scan []string
}

// Hack: Cache record of updates so we don't have to keep doing git pulls.
var updateCache = map[string]bool{}

// refFlatten recursively flattens the vendor tree.
func recFlatten(f *flattening, force bool, home string, cache, cacheGopath, useGopath bool, scanned map[string]bool) error {
	Debug("---> Inspecting %s for changes (%d packages).\n", f.curr, len(f.scan))
	for _, imp := range f.scan {
		Debug("----> Scanning %s", imp)
		base := path.Join(f.top, imp)
		mod := []string{}
		if m, ok := mergeGlide(base, imp, f); ok {
			mod = m
		} else if m, ok = mergeGodep(base, imp, f); ok {
			mod = m
		} else if m, ok = mergeGPM(base, imp, f); ok {
			mod = m
		} else if m, ok = mergeGb(base, imp, f); ok {
			mod = m
		} else if m, ok = mergeGuess(base, imp, f, scanned); ok {
			mod = m
		}

		if len(mod) > 0 {
			Debug("----> Updating all dependencies for %q (%d)", imp, len(mod))
			flattenGlideUp(f, base, home, force, cache, cacheGopath, useGopath)
			f2 := &flattening{
				conf: f.conf,
				top:  f.top,
				curr: base,
				deps: f.deps,
				scan: mod}
			recFlatten(f2, force, home, cache, cacheGopath, useGopath, scanned)
		}
	}

	return nil
}

// flattenGlideUp does a glide update in the middle of a flatten operation.
//
// While this is expensive, it is also necessary to make sure we have the
// correct version of all dependencies. We might be able to simplify by
// marking packages dirty when they are added.
func flattenGlideUp(f *flattening, base, home string, force, cache, cacheGopath, useGopath bool) error {
	//vdir := path.Join(base, "vendor")
	for _, imp := range f.deps {
		// If the top package name in the glide.yaml file is present in the deps
		// skip it because we already have it.
		if imp.Name == f.conf.Name {
			continue
		}
		wd := path.Join(f.top, imp.Name)
		if VcsExists(imp, wd) {
			if updateCache[imp.Name] {
				Debug("----> Already updated %s", imp.Name)
				continue
			}
			Debug("Updating project %s (%s)\n", imp.Name, wd)
			if err := VcsUpdate(imp, f.top, home, force, cache, cacheGopath, useGopath); err != nil {
				// We can still go on just fine even if this fails.
				Warn("Skipped update %s: %s\n", imp.Name, err)
				continue
			}
			updateCache[imp.Name] = true
		} else {
			Debug("Importing %s to project %s\n", imp.Name, wd)
			if err := VcsGet(imp, wd, home, cache, cacheGopath, useGopath); err != nil {
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

func mergeGlide(dir, name string, f *flattening) ([]string, bool) {
	deps := f.deps
	vend := f.top
	gp := path.Join(dir, "glide.yaml")
	if _, err := os.Stat(gp); err != nil {
		return []string{}, false
	}

	yml, err := ioutil.ReadFile(gp)
	if err != nil {
		Warn("Found glide file %q, but can't read: %s", gp, err)
		return []string{}, false
	}

	conf, err := cfg.ConfigFromYaml(yml)
	if err != nil {
		Warn("Found glide file %q, but can't use it: %s", gp, err)
		return []string{}, false
	}

	Info("Found glide.yaml in %s", gp)

	return mergeDeps(deps, conf.Imports, vend, f), true
}

// listGodep appends Godeps entries to the deps.
//
// It returns true if any dependencies were found (even if not added because
// they are duplicates).
func mergeGodep(dir, name string, f *flattening) ([]string, bool) {
	deps := f.deps
	vend := f.top
	Debug("Looking in %s/Godeps/ for a Godeps.json file.\n", dir)
	d, err := parseGodepGodeps(dir)
	if err != nil {
		Warn("Looking for Godeps: %s\n", err)
		return []string{}, false
	} else if len(d) == 0 {
		return []string{}, false
	}

	Info("Found Godeps.json file for %q", name)
	return mergeDeps(deps, d, vend, f), true
}

// listGb merges GB dependencies into the deps.
func mergeGb(dir, pkg string, f *flattening) ([]string, bool) {
	deps := f.deps
	vend := f.top
	Debug("Looking in %s/vendor/ for a manifest file.\n", dir)
	d, err := parseGbManifest(dir)
	if err != nil || len(d) == 0 {
		return []string{}, false
	}
	Info("Found gb manifest file for %q", pkg)
	return mergeDeps(deps, d, vend, f), true
}

// mergeGPM merges GPM Godeps files into deps.
func mergeGPM(dir, pkg string, f *flattening) ([]string, bool) {
	deps := f.deps
	vend := f.top
	d, err := parseGPMGodeps(dir)
	if err != nil || len(d) == 0 {
		return []string{}, false
	}
	Info("Found GPM file for %q", pkg)
	return mergeDeps(deps, d, vend, f), true
}

// mergeGuess guesses dependencies and merges.
//
// This always returns true because it always handles the job of searching
// for dependencies. So generally it should be the last merge strategy
// that you try.
func mergeGuess(dir, pkg string, f *flattening, scanned map[string]bool) ([]string, bool) {
	deps := f.deps
	Info("Scanning %s for dependencies.", pkg)
	buildContext, err := util.GetBuildContext()
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
	for _, oname := range d {
		if _, ok := scanned[oname]; ok {
			//Info("===> Scanned %s already. Skipping", name)
			continue
		}
		Debug("=> Scanning %s", oname)
		name, _ := NormalizeName(oname)
		//if _, ok := deps[name]; ok {
		//scanned[oname] = true
		//Debug("====> Seen %s already. Skipping", name)
		//continue
		//}
		if f.conf.HasIgnore(name) {
			Debug("==> Skipping %s because it is on the ignore list", name)
			continue
		}

		found := findPkg(buildContext, name, dir)
		switch found.PType {
		case ptypeUnknown:
			Info("==> Unknown %s (%s)", name, oname)
			Debug("✨☆ Undownloaded dependency: %s", name)
			repo := util.GetRootFromPackage(name)
			nd := &cfg.Dependency{
				Name:       name,
				Repository: "https://" + repo,
			}
			deps[name] = nd
			res = append(res, name)
		case ptypeGoroot, ptypeCgo:
			scanned[oname] = true
			// Why do we break rather than continue?
			break
		default:
			// We're looking for dependencies that might exist in $GOPATH
			// but not be on vendor. We add any that are on $GOPATH.
			if _, ok := deps[name]; !ok {
				Debug("✨☆ GOPATH dependency: %s", name)
				nd := &cfg.Dependency{Name: name}
				deps[name] = nd
				res = append(res, name)
			}
			scanned[oname] = true
		}
	}

	return res, true
}

// mergeDeps merges any dependency array into deps.
func mergeDeps(orig map[string]*cfg.Dependency, add []*cfg.Dependency, vend string, f *flattening) []string {
	mod := []string{}
	for _, dd := range add {
		if f.conf.HasIgnore(dd.Name) {
			Debug("Skipping %s because it is on the ignore list", dd.Name)
		} else if existing, ok := orig[dd.Name]; !ok {
			// Add it unless it's already there.
			orig[dd.Name] = dd
			Debug("Adding %s to the scan list", dd.Name)
			mod = append(mod, dd.Name)
		} else if existing.Reference == "" && dd.Reference != "" {
			// If a nested dep has finer dependency references than outside,
			// set the reference.
			existing.Reference = dd.Reference
			mod = append(mod, dd.Name)
		} else if dd.Reference != "" && existing.Reference != "" && dd.Reference != existing.Reference {
			// Check if one is a version and the other is a constraint. If the
			// version is in the constraint use that.
			dest := path.Join(vend, dd.Name)
			repo, err := existing.GetRepo(dest)
			if err != nil {
				Warn("Unable to access repo for %s\n", existing.Name)
				Info("Keeping %s %s", existing.Name, existing.Reference)
				continue
			}

			eIsRef := repo.IsReference(existing.Reference)
			ddIsRef := repo.IsReference(dd.Reference)

			// Both are references and different ones.
			if eIsRef && ddIsRef {
				Warn("Conflict: %s ref is %s, but also asked for %s\n", existing.Name, existing.Reference, dd.Reference)
				Info("Keeping %s %s", existing.Name, existing.Reference)
			} else if eIsRef {
				// Test ddIsRef is a constraint and if eIsRef is a semver
				// within that
				con, err := semver.NewConstraint(dd.Reference)
				if err != nil {
					Warn("Version issue for %s: '%s' is neither a reference or semantic version constraint\n", dd.Name, dd.Reference)
					Info("Keeping %s %s", existing.Name, existing.Reference)
					continue
				}

				ver, err := semver.NewVersion(existing.Reference)
				if err != nil {
					// The existing version is not a semantic version.
					Warn("Conflict: %s version is %s, but also asked for %s\n", existing.Name, existing.Reference, dd.Reference)
					Info("Keeping %s %s", existing.Name, existing.Reference)
					continue
				}

				if con.Check(ver) {
					Info("Keeping %s %s because it fits constraint '%s'", existing.Name, existing.Reference, dd.Reference)
				} else {
					Warn("Conflict: %s version is %s but does not meet constraint '%s'\n", existing.Name, existing.Reference, dd.Reference)
					Info("Keeping %s %s", existing.Name, existing.Reference)
				}

			} else if ddIsRef {
				// Test eIsRef is a constraint and if ddIsRef is a semver
				// within that
				con, err := semver.NewConstraint(existing.Reference)
				if err != nil {
					Warn("Version issue for %s: '%s' is neither a reference or semantic version constraint\n", existing.Name, existing.Reference)
					Info("Keeping %s %s", existing.Name, existing.Reference)
					continue
				}

				ver, err := semver.NewVersion(dd.Reference)
				if err != nil {
					// The dd version is not a semantic version.
					Warn("Conflict: %s version is %s, but also asked for %s\n", existing.Name, existing.Reference, dd.Reference)
					Info("Keeping %s %s", existing.Name, existing.Reference)
					continue
				}

				if con.Check(ver) {
					// Use the specific version if noted instead of the existing
					// constraint.
					existing.Reference = dd.Reference
					mod = append(mod, dd.Name)
					Info("Using %s %s because it fits constraint '%s'", existing.Name, dd.Reference, existing.Reference)
				} else {
					Warn("Conflict: %s semantic version constraint is %s but '%s' does not meet the constraint\n", existing.Name, existing.Reference, dd.Reference)
					Info("Keeping %s %s", existing.Name, existing.Reference)
				}
			} else {
				// Neither is a vcs reference and both could be semantic version
				// constraints that are different.

				_, err := semver.NewConstraint(dd.Reference)
				if err != nil {
					// dd.Reference is not a reference or a valid constraint.
					Warn("Version %s %s is not a reference or valid semantic version constraint\n", dd.Name, dd.Reference)
					Info("Keeping %s %s", existing.Name, existing.Reference)
					continue
				}

				_, err = semver.NewConstraint(existing.Reference)
				if err != nil {
					// existing.Reference is not a reference or a valid constraint.
					// We really should never end up here.
					Warn("Version %s %s is not a reference or valid semantic version constraint\n", existing.Name, existing.Reference)

					existing.Reference = dd.Reference
					mod = append(mod, dd.Name)
					Info("Using %s %s because it is a valid version", existing.Name, existing.Reference)
					continue
				}

				// Both versions are constraints. Try to merge them.
				// If either comparison has an || skip merging. That's complicated.
				ddor := strings.Index(dd.Reference, "||")
				eor := strings.Index(existing.Reference, "||")
				if ddor == -1 && eor == -1 {
					// Add the comparisons together.
					newRef := existing.Reference + ", " + dd.Reference
					existing.Reference = newRef
					mod = append(mod, dd.Name)
					Info("Combining %s semantic version constraints %s and %s", existing.Name, existing.Reference, dd.Reference)
				} else {
					Warn("Conflict: %s version is %s, but also asked for %s\n", existing.Name, existing.Reference, dd.Reference)
					Info("Keeping %s %s", existing.Name, existing.Reference)
				}
			}
		}
	}
	return mod
}
