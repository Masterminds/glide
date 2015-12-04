package cmd

import (
	"os"
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
// 	- dirname (string): Directory to use as the base. Default: "."
func GuessDeps(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	buildContext, err := GetBuildContext()
	if err != nil {
		return nil, err
	}
	base := p.Get("dirname", ".").(string)
	name := guessPackageName(buildContext, base)

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

	Info("Generating a YAML configuration file and guessing the dependencies")

	config := new(cfg.Config)
	vpath := r.VendorDir
	if !strings.HasSuffix(vpath, "/") {
		vpath = vpath + string(os.PathSeparator)
	}

	// Get the name of the top level package
	config.Name = name
	config.Imports = make([]*cfg.Dependency, len(sortable))
	i := 0
	for _, pa := range sortable {
		n := strings.TrimPrefix(pa, vpath)
		Info("Found reference to %s\n", n)
		root := util.GetRootFromPackage(n)

		d := &cfg.Dependency{
			Name: root,
		}
		subpkg := strings.TrimPrefix(n, root)
		if len(subpkg) > 0 && subpkg != "/" {
			d.Subpackages = []string{subpkg}
		}
		config.Imports[i] = d
		i++
	}

	return config, nil
}

// Attempt to guess at the package name at the top level. When unable to detect
// a name goes to default of "main".
func guessPackageName(b *BuildCtxt, base string) string {
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
