package cmd

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/Masterminds/cookoo"
	"github.com/kylelemons/go-gypsy/yaml"
)

// Recurse does glide installs on dependent packages.
// Recurse looks in all known packages for a glide.yaml files and installs for
// each one it finds.
func Recurse(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	if !p.Get("enable", true).(bool) {
		return nil, nil
	}
	force := p.Get("force", true).(bool)

	godeps, gpm := false, false
	if g, ok := p.Has("importGodeps"); ok {
		godeps = g.(bool)
	}
	if g, ok := p.Has("importGPM"); ok {
		gpm = g.(bool)
	}

	Info("Checking dependencies for updates. Godeps: %v, GPM: %v\n", godeps, gpm)
	conf := p.Get("conf", &Config{}).(*Config)
	vend, _ := VendorPath(c)

	return recDepResolve(conf, vend, godeps, gpm, force)
}

func recDepResolve(conf *Config, vend string, godeps, gpm, force bool) (interface{}, error) {

	Info("Inspecting %s.\n", vend)

	if len(conf.Imports) == 0 {
		Info("No imports.\n")
	}

	// Look in each package to see whether it has a glide.yaml, and no vendor/
	for _, imp := range conf.Imports {
		if imp.Flattened == true {
			continue
		}
		base := path.Join(vend, imp.Name)
		Info("Looking in %s for a glide.yaml file.\n", base)
		if !needsGlideUp(base) {
			if godeps {
				importGodep(base, imp.Name)
			}
			if gpm {
				importGPM(base, imp.Name)
			}
			if !needsGlideUp(base) {
				Info("Package %s manages its own dependencies.\n", imp.Name)
				continue
			}
		}

		if err := dependencyGlideUp(conf, base, godeps, gpm, force); err != nil {
			Warn("Failed to update dependency %s: %s", imp.Name, err)
		}
	}

	// Run `glide up`
	return nil, nil
}

func dependencyGlideUp(parentConf *Config, base string, godep, gpm, force bool) error {
	Info("Doing a glide in %s\n", base)
	//conf := new(Config)
	fname := path.Join(base, "glide.yaml")
	f, err := yaml.ReadFile(fname)
	if err != nil {
		return err
	}

	conf, err := FromYaml(f.Root)
	conf.Parent = parentConf
	if err != nil {
		return err
	}
	for _, imp := range conf.Imports {

		// if our root glide.yaml says to flatten this, we skip it
		if dep := conf.GetRoot().Imports.Get(imp.Name); dep != nil && dep.Flatten == true {
			Info("Skipping importing %s due to flatten being set in root glide.yaml\n", imp.Name)
			imp.Flattened = true
			continue
		}

		// We don't use the global var to find vendor dir name because the
		// user may mis-use that var to modify the local vendor dir, and
		// we don't want that to break the embedded vendor dirs.
		wd := path.Join(base, "vendor", imp.Name)
		vdir := path.Join(base, "vendor")
		if err := ensureDir(wd); err != nil {
			Warn("Skipped getting %s (vendor/ error): %s\n", imp.Name, err)
			continue
		}

		if VcsExists(imp, wd) {
			Info("Updating project %s (%s)\n", imp.Name, wd)
			if err := VcsUpdate(imp, vdir, force); err != nil {
				// We can still go on just fine even if this fails.
				Warn("Skipped update %s: %s\n", imp.Name, err)
				continue
			}
		} else {
			Info("Importing %s to project %s\n", imp.Name, base)
			if err := VcsGet(imp, wd); err != nil {
				Warn("Skipped getting %s: %v\n", imp.Name, err)
				continue
			}
		}

		// If a revision has been set use it.
		err = VcsVersion(imp, vdir)
		if err != nil {
			Warn("Problem setting version on %s: %s\n", imp.Name, err)
		}

		//recDepResolve(conf, path.Join(wd, "vendor"))
	}
	recDepResolve(conf, path.Join(base, "vendor"), godep, gpm, force)
	return nil
}

func ensureDir(dirpath string) error {
	if fi, err := os.Stat(dirpath); err == nil && fi.IsDir() {
		return nil
	}
	return os.MkdirAll(dirpath, 0755)
}

func needsGlideUp(dir string) bool {
	stat, err := os.Stat(path.Join(dir, "glide.yaml"))
	if err != nil || stat.IsDir() {
		return false
	}

	// Should probably see if vendor is there and non-empty.

	return true
}

func importGodep(dir, pkg string) error {
	Info("Looking in %s/Godeps/ for a Godeps.json file.\n", dir)
	d, err := parseGodepGodeps(dir)
	if err != nil {
		Warn("Looking for Godeps: %s\n", err)
		return err
	}
	return quickDirtyYAMLWrite(dir, d, pkg)
}

func importGPM(dir, pkg string) error {
	d, err := parseGPMGodeps(dir)
	if err != nil {
		return err
	}
	return quickDirtyYAMLWrite(dir, d, pkg)
}

func quickDirtyYAMLWrite(dir string, d []*Dependency, pkg string) error {
	if len(d) == 0 {
		return nil
	}
	c := &Config{Name: pkg, Imports: d}
	node := c.ToYaml()
	data := yaml.Render(node)
	f := path.Join(dir, "glide.yaml")
	Info("Writing new glide.yaml file in %s\n", dir)
	return ioutil.WriteFile(f, []byte(data), 0755)
}
