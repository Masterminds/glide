package cmd

import (
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/msg"
)

// Tree prints a tree representing dependencies.
func Tree(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	buildContext, err := GetBuildContext()
	if err != nil {
		return nil, err
	}
	showcore := p.Get("showcore", false).(bool)
	basedir := p.Get("dir", ".").(string)
	myName := guessPackageName(buildContext, basedir)

	if basedir == "." {
		var err error
		basedir, err = os.Getwd()
		if err != nil {
			Error("Could not get working directory")
			return nil, err
		}
	}

	fmt.Println(myName)
	l := list.New()
	l.PushBack(myName)
	displayTree(buildContext, basedir, myName, 1, showcore, l)
	return nil, nil
}

// ListDeps lists all of the dependencies of the current project.
//
// Params:
//  - dir (string): basedir
//  - deep (bool): whether to do a deep scan or a shallow scan
//
// Returns:
//
func ListDeps(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	basedir := p.Get("dir", ".").(string)
	deep := p.Get("deep", true).(bool)

	basedir, err := filepath.Abs(basedir)
	if err != nil {
		return nil, err
	}

	r, err := dependency.NewResolver(basedir)
	if err != nil {
		return nil, err
	}
	h := &dependency.DefaultMissingPackageHandler{Missing: []string{}, Gopath: []string{}}
	r.Handler = h

	sortable, err := r.ResolveLocal(deep)
	if err != nil {
		return nil, err
	}

	sort.Strings(sortable)

	fmt.Println("INSTALLED packages:")
	for _, k := range sortable {
		v, err := filepath.Rel(basedir, k)
		if err != nil {
			msg.Warn("Failed to Rel path: %s", err)
			v = k
		}
		fmt.Printf("\t%s\n", v)
	}

	if len(h.Missing) > 0 {
		fmt.Println("\nMISSING packages:")
		for _, pkg := range h.Missing {
			fmt.Printf("\t%s\n", pkg)
		}
	}
	if len(h.Gopath) > 0 {
		fmt.Println("\nGOPATH packages:")
		for _, pkg := range h.Gopath {
			fmt.Printf("\t%s\n", pkg)
		}
	}

	return nil, nil
}

func listDeps(b *BuildCtxt, info map[string]*pinfo, name, path string) {
	found := findPkg(b, name, path)
	switch found.PType {
	case ptypeUnknown:
		info[name] = found
		break
	case ptypeGoroot, ptypeCgo:
		break
	default:
		info[name] = found
		for _, i := range walkDeps(b, found.Path, found.Name) {
			// Only walk the deps that are not already found to avoid
			// infinite recursion.
			if _, f := info[found.Name]; f == false {
				listDeps(b, info, i, found.Path)
			}
		}
	}
}

func displayTree(b *BuildCtxt, basedir, myName string, level int, core bool, l *list.List) {
	deps := walkDeps(b, basedir, myName)
	for _, name := range deps {
		found := findPkg(b, name, basedir)
		if found.PType == ptypeUnknown {
			msg := "glide get " + found.Name
			fmt.Printf("\t%s\t(%s)\n", found.Name, msg)
			continue
		}
		if !core && found.PType == ptypeGoroot || found.PType == ptypeCgo {
			continue
		}
		fmt.Print(strings.Repeat("\t", level))

		f := findInList(found.Name, l)
		if f == true {
			fmt.Printf("(Recursion) %s   (%s)\n", found.Name, found.Path)
		} else {
			// Every branch in the tree is a copy to handle all the branches
			cl := copyList(l)
			cl.PushBack(found.Name)
			fmt.Printf("%s   (%s)\n", found.Name, found.Path)
			displayTree(b, found.Path, found.Name, level+1, core, cl)
		}
	}
}

type ptype int8

const (
	ptypeUnknown ptype = iota
	ptypeLocal
	ptypeVendor
	ptypeGopath
	ptypeGoroot
	ptypeCgo
)

func ptypeString(t ptype) string {
	switch t {
	case ptypeLocal:
		return "local"
	case ptypeVendor:
		return "vendored"
	case ptypeGopath:
		return "gopath"
	case ptypeGoroot:
		return "core"
	case ptypeCgo:
		return "cgo"
	default:
		return "missing"
	}
}

type pinfo struct {
	Name, Path string
	PType      ptype
	Vendored   bool
}

func findPkg(b *BuildCtxt, name, cwd string) *pinfo {
	var fi os.FileInfo
	var err error
	var p string

	info := &pinfo{
		Name: name,
	}

	// Recurse backward to scan other vendor/ directories
	// If the cwd isn't an absolute path walking upwards looking for vendor/
	// folders can get into an infinate loop.
	abs, err := filepath.Abs(cwd)
	if err != nil {
		abs = cwd
	}
	if abs != "." {
		for wd := abs; wd != "/"; wd = filepath.Dir(wd) {
			p = filepath.Join(wd, "vendor", name)
			if fi, err = os.Stat(p); err == nil && (fi.IsDir() || isLink(fi)) {
				info.Path = p
				info.PType = ptypeVendor
				info.Vendored = true
				return info
			}
		}
	}
	// Check $GOPATH
	for _, r := range strings.Split(b.GOPATH, ":") {
		p = filepath.Join(r, "src", name)
		if fi, err = os.Stat(p); err == nil && (fi.IsDir() || isLink(fi)) {
			info.Path = p
			info.PType = ptypeGopath
			return info
		}
	}

	// Check $GOROOT
	for _, r := range strings.Split(b.GOROOT, ":") {
		p = filepath.Join(r, "src", name)
		if fi, err = os.Stat(p); err == nil && (fi.IsDir() || isLink(fi)) {
			info.Path = p
			info.PType = ptypeGoroot
			return info
		}
	}

	// Finally, if this is "C", we're dealing with cgo
	if name == "C" {
		info.PType = ptypeCgo
	}

	return info
}

func isLink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink == os.ModeSymlink
}

func walkDeps(b *BuildCtxt, base, myName string) []string {
	externalDeps := []string{}
	filepath.Walk(base, func(path string, fi os.FileInfo, err error) error {
		if excludeSubtree(path, fi) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		pkg, err := b.ImportDir(path, 0)
		if err != nil {
			if !strings.HasPrefix(err.Error(), "no buildable Go source") {
				Warn("Error: %s (%s)", err, path)
				// Not sure if we should return here.
				//return err
			}
		}

		if pkg.Goroot {
			return nil
		}

		for _, imp := range pkg.Imports {
			//if strings.HasPrefix(imp, myName) {
			////Info("Skipping %s because it is a subpackage of %s", imp, myName)
			//continue
			//}
			if imp == myName {
				continue
			}
			externalDeps = append(externalDeps, imp)
		}

		return nil
	})
	return externalDeps
}

func excludeSubtree(path string, fi os.FileInfo) bool {
	top := filepath.Base(path)

	if !fi.IsDir() && !isLink(fi) {
		return true
	}

	// Provisionally, we'll skip vendor. We definitely
	// should skip testdata.
	if top == "vendor" || top == "testdata" {
		return true
	}

	// Skip anything that starts with _
	if strings.HasPrefix(top, "_") || (strings.HasPrefix(top, ".") && top != ".") {
		return true
	}
	return false
}

func copyList(l *list.List) *list.List {
	n := list.New()
	for e := l.Front(); e != nil; e = e.Next() {
		n.PushBack(e.Value.(string))
	}
	return n
}

func findInList(n string, l *list.List) bool {
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value.(string) == n {
			return true
		}
	}

	return false
}
