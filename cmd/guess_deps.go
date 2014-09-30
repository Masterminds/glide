package cmd

import (
	"bytes"
	"github.com/Masterminds/cookoo"
	"go/build"
	"os"
	"text/template"
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
	delete(deps, base)
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("main").Parse(yamlGuessTpl)
	if err != nil {
		return nil, err
	}
	var doc bytes.Buffer
	tmpl.Execute(&doc, deps)
	Info(doc.String())
	return doc, nil
}

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
