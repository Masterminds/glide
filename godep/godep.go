// Package godep provides basic importing of Godep dependencies.
//
// This is not a complete implementation of Godep.
package godep

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
)

// This file contains commands for working with Godep.

// The Godeps struct from Godep.
//
// https://raw.githubusercontent.com/tools/godep/master/dep.go
//
// We had to copy this because it's in the package main for Godep.
type Godeps struct {
	ImportPath string
	GoVersion  string
	Packages   []string `json:",omitempty"` // Arguments to save, if any.
	Deps       []Dependency

	outerRoot string
}

// Dependency is a modified version of Godep's Dependency struct.
// It drops all of the unexported fields.
type Dependency struct {
	ImportPath string
	Comment    string `json:",omitempty"` // Description of commit, if present.
	Rev        string // VCS-specific commit ID.
}

// Has is a command to detect if a package contains a Godeps.json file.
func Has(dir string) bool {
	path := filepath.Join(dir, "Godeps/Godeps.json")
	_, err := os.Stat(path)
	return err == nil
}

// Parse parses a Godep's Godeps file.
//
// It returns the contents as a dependency array.
func Parse(dir string) ([]*cfg.Dependency, error) {
	path := filepath.Join(dir, "Godeps/Godeps.json")
	if _, err := os.Stat(path); err != nil {
		return []*cfg.Dependency{}, nil
	}
	msg.Info("Found Godeps.json file in %s", gpath.StripBasepath(dir))
	msg.Info("--> Parsing Godeps metadata...")

	buf := []*cfg.Dependency{}

	godeps := &Godeps{}

	// Get a handle to the file.
	file, err := os.Open(path)
	if err != nil {
		return buf, err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	if err := dec.Decode(godeps); err != nil {
		return buf, err
	}

	seen := map[string]bool{}
	for _, d := range godeps.Deps {
		pkg, sub := util.NormalizeName(d.ImportPath)
		if _, ok := seen[pkg]; ok {
			if len(sub) == 0 {
				continue
			}
			// Modify existing dep with additional subpackages.
			for _, dep := range buf {
				if dep.Name == pkg {
					dep.Subpackages = append(dep.Subpackages, sub)
				}
			}
		} else {
			seen[pkg] = true
			dep := &cfg.Dependency{Name: pkg, Reference: d.Rev}
			if sub != "" {
				dep.Subpackages = []string{sub}
			}
			buf = append(buf, dep)
		}
	}

	return buf, nil
}

// RemoveGodepSubpackages strips subpackages from a cfg.Config dependencies that
// contain "Godeps/_workspace/src" as part of the path.
func RemoveGodepSubpackages(c *cfg.Config) *cfg.Config {
	for _, d := range c.Imports {
		n := []string{}
		for _, v := range d.Subpackages {
			if !strings.HasPrefix(v, "Godeps/_workspace/src") {
				n = append(n, v)
			}
		}
		d.Subpackages = n
	}

	for _, d := range c.DevImports {
		n := []string{}
		for _, v := range d.Subpackages {
			if !strings.HasPrefix(v, "Godeps/_workspace/src") {
				n = append(n, v)
			}
		}
		d.Subpackages = n
	}

	return c
}
