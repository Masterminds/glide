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
func List(basedir string, deep bool) {

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

	sortable, err := r.ResolveLocal(deep)
	if err != nil {
		msg.Die("Error listing dependencies: %s", err)
	}

	sort.Strings(sortable)

	msg.Puts("INSTALLED packages:")
	for _, k := range sortable {
		v, err := filepath.Rel(basedir, k)
		if err != nil {
			msg.Warn("Failed to Rel path: %s", err)
			v = k
		}
		msg.Puts("\t%s", v)
	}

	if len(h.Missing) > 0 {
		msg.Puts("\nMISSING packages:")
		for _, pkg := range h.Missing {
			msg.Puts("\t%s", pkg)
		}
	}
	if len(h.Gopath) > 0 {
		msg.Puts("\nGOPATH packages:")
		for _, pkg := range h.Gopath {
			msg.Puts("\t%s", pkg)
		}
	}
}
