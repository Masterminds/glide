package action

import (
	"path/filepath"
	"sort"

	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/msg"
)

// List lists all of the dependencies of the current project.
//
// Params:
//  - dir (string): basedir
//  - deep (bool): whether to do a deep scan or a shallow scan
func List(basedir string, deep bool) PackageList {
	basedir, err := filepath.Abs(basedir)
	if err != nil {
		msg.Die("Could not read directory: %s", err)
	}

	r, err := dependency.NewResolver(basedir)
	if err != nil {
		msg.Die("Could not create a resolver: %s", err)
	}
	h := &dependency.DefaultMissingPackageHandler{Missing: []string{}, Gopath: []string{}}
	r.Handler = h

	localPkgs, err := r.ResolveLocal(deep)
	if err != nil {
		msg.Die("Error listing dependencies: %s", err)
	}
	sort.Strings(localPkgs)
	installed := make([]string, len(localPkgs))
	for i, pkg := range localPkgs {
		relPkg, err := filepath.Rel(basedir, pkg)
		if err != nil {
			// msg.Warn("Failed to Rel path: %s", err)
			relPkg = pkg
		}
		installed[i] = relPkg
	}
	return PackageList{
		Installed: installed,
		Missing:   h.Missing,
		Gopath:    h.Gopath,
	}
}

type PackageList struct {
	Installed []string `json:"installed"`
	Missing   []string `json:"missing"`
	Gopath    []string `json:"gopath"`
}
