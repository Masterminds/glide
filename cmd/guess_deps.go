package cmd

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/util"
)

// GuessDeps tries to get the dependencies for the current directory.
//
// Params
//  - dirname (string): Directory to use as the base. Default: "."
//  - skipImport (book): Whether to skip importing from Godep, GPM, and gb
func GuessDeps(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	buildContext, err := util.GetBuildContext()
	if err != nil {
		return nil, err
	}
	base := p.Get("dirname", ".").(string)
	skipImport := p.Get("skipImport", false).(bool)
	name := guessPackageName(buildContext, base)

	Info("Generating a YAML configuration file and guessing the dependencies")

	config := new(cfg.Config)

	// Get the name of the top level package
	config.Name = name

	// Import by looking at other package managers and looking over the
	// entire directory structure.

	// Attempt to import from other package managers.
	if !skipImport {
		Info("Attempting to import from other package managers (use --skip-import to skip)")
		deps := []*cfg.Dependency{}
		absBase, err := filepath.Abs(base)
		if err != nil {
			return nil, err
		}

		if d, ok := guessImportGodep(absBase); ok {
			Info("Importing Godep configuration")
			Warn("Godep uses commit id versions. Consider using Semantic Versions with Glide")
			deps = d
		} else if d, ok := guessImportGPM(absBase); ok {
			Info("Importing GPM configuration")
			deps = d
		} else if d, ok := guessImportGB(absBase); ok {
			Info("Importing GB configuration")
			deps = d
		}

		for _, i := range deps {
			Info("Found imported reference to %s\n", i.Name)
			config.Imports = append(config.Imports, i)
		}
	}

	// Resolve dependencies by looking at the tree.
	r, err := dependency.NewResolver(base)
	if err != nil {
		return nil, err
	}

	h := &dependency.DefaultMissingPackageHandler{Missing: []string{}, Gopath: []string{}}
	r.Handler = h

	sortable, err := r.ResolveLocal(false)
	if err != nil {
		return nil, err
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
			Info("Found reference to %s\n", n)
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
					Info("Adding sub-package %s to %s\n", subpkg, root)
					d.Subpackages = append(d.Subpackages, subpkg)
				}
			}
		}
	}

	return config, nil
}

// Attempt to guess at the package name at the top level. When unable to detect
// a name goes to default of "main".
//
// DEPRECATED. Use BuildCtxt.PackageName(base)
func guessPackageName(b *util.BuildCtxt, base string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return "main"
	}

	pkg, err := b.Import(base, cwd, 0)
	if err != nil {
		// There may not be any top level Go source files but the project may
		// still be within the GOPATH.
		if strings.HasPrefix(base, b.GOPATH) {
			p := strings.TrimPrefix(base, b.GOPATH)
			return strings.Trim(p, string(os.PathSeparator))
		}
	}

	return pkg.ImportPath
}

func guessImportGodep(dir string) ([]*cfg.Dependency, bool) {
	d, err := parseGodepGodeps(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}

func guessImportGPM(dir string) ([]*cfg.Dependency, bool) {
	d, err := parseGPMGodeps(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}

func guessImportGB(dir string) ([]*cfg.Dependency, bool) {
	d, err := parseGbManifest(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}
