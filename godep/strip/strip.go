// Package strip removes Godeps/_workspace and undoes the Godep rewrites. This
// essentially removes the old style (pre-vendor) Godep vendoring.
//
// Note, this functionality is deprecated. Once more projects use the Godep
// support for the core vendoring this will no longer be needed.
package strip

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Masterminds/glide/msg"
)

var godepMark = map[string]bool{}

var vPath = "vendor"

// GodepWorkspace removes any Godeps/_workspace directories and makes sure
// any rewrites are undone.
// Note, this is not concuccency safe.
func GodepWorkspace(v string) error {
	vPath = v
	if _, err := os.Stat(vPath); err != nil {
		if os.IsNotExist(err) {
			msg.Debug("Vendor directory does not exist.")
		}

		return err
	}

	err := filepath.Walk(vPath, stripGodepWorkspaceHandler)
	if err != nil {
		return err
	}

	// Walk the marked projects to make sure rewrites are undone.
	for k := range godepMark {
		msg.Info("Removing Godep rewrites for %s", k)
		err := filepath.Walk(k, rewriteGodepfilesHandler)
		if err != nil {
			return err
		}
	}

	return nil
}

func stripGodepWorkspaceHandler(path string, info os.FileInfo, err error) error {
	// Skip the base vendor directory
	if path == vPath {
		return nil
	}

	name := info.Name()
	p := filepath.Dir(path)
	pn := filepath.Base(p)
	if name == "_workspace" && pn == "Godeps" {
		if _, err := os.Stat(path); err == nil {
			if info.IsDir() {
				// Marking this location to make sure rewrites are undone.
				pp := filepath.Dir(p)
				godepMark[pp] = true

				msg.Info("Removing: %s", path)
				return os.RemoveAll(path)
			}

			msg.Debug("%s is not a directory. Skipping removal", path)
			return nil
		}
	}
	return nil
}

func rewriteGodepfilesHandler(path string, info os.FileInfo, err error) error {
	name := info.Name()
	if name == "testdata" || name == "vendor" {
		return filepath.SkipDir
	}

	if info.IsDir() {
		return nil
	}

	if e := filepath.Ext(path); e != ".go" {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	var changed bool
	for _, s := range f.Imports {
		n, err := strconv.Unquote(s.Path.Value)
		if err != nil {
			return err
		}
		q := rewriteGodepImport(n)
		if q != name {
			s.Path.Value = strconv.Quote(q)
			changed = true
		}
	}
	if !changed {
		return nil
	}

	printerConfig := &printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}
	var buffer bytes.Buffer
	if err = printerConfig.Fprint(&buffer, fset, f); err != nil {
		return err
	}
	fset = token.NewFileSet()
	f, err = parser.ParseFile(fset, name, &buffer, parser.ParseComments)
	ast.SortImports(fset, f)
	tpath := path + ".temp"
	t, err := os.Create(tpath)
	if err != nil {
		return err
	}
	if err = printerConfig.Fprint(t, fset, f); err != nil {
		return err
	}
	if err = t.Close(); err != nil {
		return err
	}

	msg.Debug("Rewriting Godep imports for %s", path)

	// This is required before the rename on windows.
	if err = os.Remove(path); err != nil {
		return err
	}
	return os.Rename(tpath, path)
}

func rewriteGodepImport(n string) string {
	if !strings.Contains(n, "Godeps/_workspace/src") {
		return n
	}

	i := strings.LastIndex(n, "Godeps/_workspace/src")

	return strings.TrimPrefix(n[i:], "Godeps/_workspace/src/")
}
