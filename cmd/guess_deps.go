package cmd

import (
	"github.com/Masterminds/cookoo"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"sort"
	"os"
)

// GuessDeps tries to get the dependencies for the current directory.
//
// This scans the given directory and its immediate subdirectories and tries
// to guess all of the package dependencies.
//
// TODO: Walk all the way down the directory tree.
//
// Params
// 	- dirname (string): Directory to use as the base. Default: "."
func GuessDeps(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	base := p.Get("dirname", ".").(string)

	f, err := os.Open(base)
	if err != nil {
		return nil, err
	}

	// MPB: Admittedly, 1024 is an arbitrary number.
	dirs, err := f.Readdirnames(1024)
	dirs = append(dirs, base)
	if err != nil {
		return nil, err
	}

	includes := make(map[string]bool, 10)
	for _, d := range dirs {
		imps, err := importsInDir(d)
		if err != nil {
			continue
		}

		// From imports, get candidates for VCS paths.
		for _, i := range imps {
			i = vcsPath(i)
			includes[i] = true
		}
	}


	norminc := normalizeIncludes(includes)
	Info("%v\n", norminc)

	return norminc, err
}

// importsInDir finds the imports in all .go files in a directory.
func importsInDir(dir string) ([]string, error) {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, dir, nil, parser.ImportsOnly)
	if err != nil {
		return []string{}, err
	}

	buf := []string{}
	for _, pkg := range pkgs {
		Info("Scanning %s for imports.\n", pkg.Name)
		for _, f := range pkg.Files {
			for _, imp := range f.Imports {
				p, _ := strconv.Unquote(imp.Path.Value)
				// Basically, anything with less than 2 slashes is
				// ungettable to 'go get' or our VCS versions.
				if strings.Count(p, "/") > 1 {
					buf = append(buf, p)
				}
			}
		}
	}

	return buf, nil
}

// vcsPath gets the VCS path for a given import.
//
// Not totally sure that this method matches 'go get'
func vcsPath(imp string) string {
	parts := strings.SplitN(imp, "/", 4)
	if len(parts) < 4 {
		return imp
	}

	repo := strings.Join(parts[:3], "/")
	Info("Normalized %s to %s\n", imp, repo)
	return repo
}

// normalizeIncludes takes a map of includes and returns a lexigraphically sorted array.
func normalizeIncludes(incs map[string]bool) []string {
	buf := make([]string, len(incs))
	i := 0
	for k, _ := range incs {
		buf[i] = k
		i++
	}
	sort.Strings(buf)
	return buf
}
