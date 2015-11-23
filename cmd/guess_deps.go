package cmd

import (
	"os"
	"strings"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/cfg"
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
	deps := make(map[string]bool)
	err = findDeps(buildContext, deps, base, "")
	deps = compactDeps(deps)
	delete(deps, base)
	if err != nil {
		return nil, err
	}

	Info("Generating a YAML configuration file and guessing the dependencies")

	config := new(cfg.Config)

	// Get the name of the top level package
	config.Name = guessPackageName(buildContext, base)
	config.Imports = make([]*cfg.Dependency, len(deps))
	i := 0
	for pa := range deps {
		Info("Found reference to %s\n", pa)
		d := &cfg.Dependency{
			Name: pa,
		}
		config.Imports[i] = d
		i++
	}

	return config, nil
}

// findDeps finds all of the dependenices.
// https://golang.org/src/cmd/go/pkg.go#485
//
// As of Go 1.5 the go command knows about the vendor directory but the go/build
// package does not. It only knows about the GOPATH and GOROOT. In order to look
// for packages in the vendor/ directory we need to fake it for now.
func findDeps(b *BuildCtxt, soFar map[string]bool, name, vpath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Skip cgo pseudo-package.
	if name == "C" {
		return nil
	}

	pkg, err := b.Import(name, cwd, 0)
	if err != nil {
		return err
	}

	if pkg.Goroot {
		return nil
	}

	if vpath == "" {
		vpath = pkg.ImportPath
	}

	// When the vendor/ directory is present make sure we strip it out before
	// registering it as a guess.
	realName := strings.TrimPrefix(pkg.ImportPath, vpath+"/vendor/")

	// Before adding a name to the list make sure it's not the name of the
	// top level package.
	lookupName, _ := NormalizeName(realName)
	if vpath != lookupName {
		soFar[realName] = true
	}
	for _, imp := range pkg.Imports {
		if !soFar[imp] {

			// Try looking for a dependency as a vendor. If it's not there then
			// fall back to a way where it will be found in the GOPATH or GOROOT.
			if err := findDeps(b, soFar, vpath+"/vendor/"+imp, vpath); err != nil {
				if err := findDeps(b, soFar, imp, vpath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// compactDeps registers only top level packages.
//
// Minimize the package imports. For example, importing github.com/Masterminds/cookoo
// and github.com/Masterminds/cookoo/io should not import two packages. Only one
// package needs to be referenced.
func compactDeps(soFar map[string]bool) map[string]bool {
	basePackages := make(map[string]bool, len(soFar))
	for k := range soFar {
		base, _ := NormalizeName(k)
		basePackages[base] = true
	}

	return basePackages
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
		return "main"
	}

	return pkg.ImportPath
}
