package action

import (
	"encoding/json"
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
//  - format (string): The format to output (text, json, json-pretty)
func List(basedir string, deep bool, format string) {
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
	l := PackageList{
		Installed: installed,
		Missing:   h.Missing,
		Gopath:    h.Gopath,
	}

	outputList(l, format)
}

// PackageList contains the packages being used by their location
type PackageList struct {
	Installed []string `json:"installed"`
	Missing   []string `json:"missing"`
	Gopath    []string `json:"gopath"`
}

const (
	textFormat       = "text"
	jsonFormat       = "json"
	jsonPrettyFormat = "json-pretty"
)

func outputList(l PackageList, format string) {
	switch format {
	case textFormat:
		msg.Puts("INSTALLED packages:")
		for _, pkg := range l.Installed {
			msg.Puts("\t%s", pkg)
		}

		if len(l.Missing) > 0 {
			msg.Puts("\nMISSING packages:")
			for _, pkg := range l.Missing {
				msg.Puts("\t%s", pkg)
			}
		}
		if len(l.Gopath) > 0 {
			msg.Puts("\nGOPATH packages:")
			for _, pkg := range l.Gopath {
				msg.Puts("\t%s", pkg)
			}
		}
	case jsonFormat:
		json.NewEncoder(msg.Default.Stdout).Encode(l)
	case jsonPrettyFormat:
		b, err := json.MarshalIndent(l, "", "  ")
		if err != nil {
			msg.Die("could not unmarshal package list: %s", err)
		}
		msg.Puts(string(b))
	default:
		msg.Die("invalid output format: must be one of: json|json-pretty|text")
	}
}
