package cmd

import (
	"go/build"
	"os"

	"github.com/Masterminds/cookoo"
)

var yamlGuessTpl = `
# Detected project's dependencies.
import:{{range $path, $notLocal := .}}
  - package: {{$path}}{{end}}
`

// GuessDeps tries to get the dependencies for the current directory.
//
// Params
// 	- dirname (string): Directory to use as the base. Default: "."
func GuessDeps(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	base := p.Get("dirname", ".").(string)
	deps := make(map[string]bool)
	err := findDeps(deps, base)
	deps = compactDeps(deps)
	delete(deps, base)
	if err != nil {
		return nil, err
	}

	config := new(Config)
	config.Imports = make([]*Dependency, len(deps))
	i := 0
	for p, _ := range deps {
		Info("Found reference to %s\n", p)
		d := &Dependency{
			Name: p,
		}
		config.Imports[i] = d
		i++
	}

	return config, nil
}

// findDeps finds all of the dependenices.
// https://golang.org/src/cmd/go/pkg.go#485
func findDeps(soFar map[string]bool, name string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	pkg, err := build.Import(name, cwd, 0)
	if err != nil {
		return err
	}

	if pkg.Goroot {
		return nil
	}

	soFar[pkg.ImportPath] = true
	for _, imp := range pkg.Imports {
		if !soFar[imp] {
			if err := findDeps(soFar, imp); err != nil {
				return err
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
	for k, _ := range soFar {
		base, _ := NormalizeName(k)
		basePackages[base] = true
	}

	return basePackages
}
