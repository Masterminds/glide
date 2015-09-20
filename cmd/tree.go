package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/cookoo"
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
	displayTree(buildContext, basedir, myName, 1, showcore)
	return nil, nil
}

// ListDeps lists all of the dependencies of the current project.
//
// Params:
//
// Returns:
//
func ListDeps(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	buildContext, err := GetBuildContext()
	if err != nil {
		return nil, err
	}
	basedir := p.Get("dir", ".").(string)
	myName := guessPackageName(buildContext, basedir)

	basedir, err = filepath.Abs(basedir)
	if err != nil {
		return nil, err
	}

	direct := map[string]bool{}
	d := walkDeps(buildContext, basedir, myName)
	for _, i := range d {
		listDeps(buildContext, direct, i, basedir)
	}

	sortable := make([]string, len(direct))
	i := 0
	for k := range direct {
		sortable[i] = k
		i++
	}

	sort.Strings(sortable)

	for _, k := range sortable {
		dec := "no"
		if direct[k] {
			dec = "yes"
		}
		fmt.Printf("%s (Present: %s)\n", k, dec)
	}

	return nil, nil
}

func listDeps(b *BuildCtxt, info map[string]bool, name, path string) {
	found := findPkg(b, name, path)
	switch found.PType {
	case ptypeUnknown:
		info[name] = false
		break
	case ptypeGoroot, ptypeCgo:
		break
	default:
		info[name] = true
		for _, i := range walkDeps(b, found.Path, found.Name) {
			listDeps(b, info, i, found.Path)
		}
	}
}

func displayTree(b *BuildCtxt, basedir, myName string, level int, core bool) {
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
		fmt.Printf("%s   (%s)\n", found.Name, found.Path)
		displayTree(b, found.Path, found.Name, level+1, core)
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

type pinfo struct {
	Name, Path string
	PType      ptype
}

func findPkg(b *BuildCtxt, name, cwd string) *pinfo {
	var fi os.FileInfo
	var err error
	var p string

	info := &pinfo{
		Name: name,
	}

	// Recurse backward to scan other vendor/ directories
	for wd := cwd; wd != "/"; wd = filepath.Dir(wd) {
		p = filepath.Join(wd, "vendor", name)
		if fi, err = os.Stat(p); err == nil && (fi.IsDir() || isLink(fi)) {
			info.Path = p
			info.PType = ptypeVendor
			return info
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
			return err
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
