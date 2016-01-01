package action

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/gb"
	"github.com/Masterminds/glide/godep"
	"github.com/Masterminds/glide/gpm"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
)

func Create(base string, skipImport bool) {
	glidefile := gpath.GlideFile
	// Guard against overwrites.
	guardYAML(glidefile)

	// Guess deps
	conf := guessDeps(base, skipImport)
	// Write YAML
	buf, err := conf.Marshal()
	if err != nil {
		msg.Die("Could not marshal config file: %s", err)
	}

	if err := ioutil.WriteFile(glidefile, buf, 0666); err != nil {
		msg.Die("Could not save %s: %s", glidefile, err)
	}
}

// guardYAML fails if the given file already exists.
//
// This prevents an important file from being overwritten.
func guardYAML(filename string) {
	if _, err := os.Stat(filename); err == nil {
		msg.Die("Cowardly refusing to overwrite existing YAML.")
	}
}

// guessDeps attempts to resolve all of the dependencies for a given project.
//
// base is the directory to start with.
// skipImport will skip running the automatic imports.
//
// FIXME: This function is likely a one-off that has a more standard alternative.
// It's also long and could use a refactor.
func guessDeps(base string, skipImport bool) *cfg.Config {
	buildContext, err := util.GetBuildContext()
	if err != nil {
		msg.Die("Failed to build an import context: %s", err)
	}
	name := buildContext.PackageName(base)

	msg.Info("Generating a YAML configuration file and guessing the dependencies")

	config := new(cfg.Config)

	// Get the name of the top level package
	config.Name = name

	// Import by looking at other package managers and looking over the
	// entire directory structure.

	// Attempt to import from other package managers.
	if !skipImport {
		msg.Info("Attempting to import from other package managers (use --skip-import to skip)")
		deps := []*cfg.Dependency{}
		absBase, err := filepath.Abs(base)
		if err != nil {
			msg.Die("Failed to resolve location of %s: %s", base, err)
		}

		if d, ok := guessImportGodep(absBase); ok {
			msg.Info("Importing Godep configuration")
			msg.Warn("Godep uses commit id versions. Consider using Semantic Versions with Glide")
			deps = d
		} else if d, ok := guessImportGPM(absBase); ok {
			msg.Info("Importing GPM configuration")
			deps = d
		} else if d, ok := guessImportGB(absBase); ok {
			msg.Info("Importing GB configuration")
			deps = d
		}

		for _, i := range deps {
			msg.Info("Found imported reference to %s\n", i.Name)
			config.Imports = append(config.Imports, i)
		}
	}

	// Resolve dependencies by looking at the tree.
	r, err := dependency.NewResolver(base)
	if err != nil {
		msg.Die("Error creating a dependency resolver: %s", err)
	}

	h := &dependency.DefaultMissingPackageHandler{Missing: []string{}, Gopath: []string{}}
	r.Handler = h

	sortable, err := r.ResolveLocal(false)
	if err != nil {
		msg.Die("Error resolving local dependencies: %s", err)
	}

	sort.Strings(sortable)

	vpath := r.VendorDir
	if !strings.HasSuffix(vpath, "/") {
		vpath = vpath + string(os.PathSeparator)
	}

	for _, pa := range sortable {
		n := strings.TrimPrefix(pa, vpath)
		root := util.GetRootFromPackage(n)

		if !config.HasDependency(root) {
			msg.Info("Found reference to %s\n", n)
			d := &cfg.Dependency{
				Name: root,
			}
			subpkg := strings.TrimPrefix(n, root)
			if len(subpkg) > 0 && subpkg != "/" {
				d.Subpackages = []string{subpkg}
			}
			config.Imports = append(config.Imports, d)
		} else {
			subpkg := strings.TrimPrefix(n, root)
			if len(subpkg) > 0 && subpkg != "/" {
				subpkg = strings.TrimPrefix(subpkg, "/")
				d := config.Imports.Get(root)
				f := false
				for _, v := range d.Subpackages {
					if v == subpkg {
						f = true
					}
				}
				if !f {
					msg.Info("Adding sub-package %s to %s\n", subpkg, root)
					d.Subpackages = append(d.Subpackages, subpkg)
				}
			}
		}
	}

	return config
}

func guessImportGodep(dir string) ([]*cfg.Dependency, bool) {
	d, err := godep.Parse(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}

func guessImportGPM(dir string) ([]*cfg.Dependency, bool) {
	d, err := gpm.Parse(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}

func guessImportGB(dir string) ([]*cfg.Dependency, bool) {
	d, err := gb.Parse(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}
