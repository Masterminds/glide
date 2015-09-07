package cmd

import (
	"errors"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/cookoo"
)

var ErrPkgNotFound = errors.New("package not found")

// Tree prints a tree representing dependencies.
func Tree(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	showcore := p.Get("showcore", false).(bool)
	basedir := p.Get("dir", ".").(string)
	myName := guessPackageName(basedir)

	if basedir == "." {
		var err error
		basedir, err = os.Getwd()
		if err != nil {
			Error("Could not get working directory")
			return nil, err
		}
	}

	fmt.Println(myName)

	// Start with *.go
	// Do a breadth-first search of subdirectories, excluding _*, .*, testdata,
	// and vendor.
	/*
		deps := walkDeps(basedir, myName)
		for _, name := range deps {
			found := findPkg(name, basedir)
			msg := found.Path
			if found.PType == ptypeUnknown {
				msg = "glide get " + found.Name
			}
			if !showcore && found.PType == ptypeGoroot {
				continue
			}
			fmt.Printf("\t%s\t(%s)\n", found.Name, msg)
		}
	*/
	displayTree(basedir, myName, 1, showcore)

	// Now look up all of the external dependencies.

	// Make it pretty.
	return nil, nil
}

func displayTree(basedir, myName string, level int, core bool) {
	deps := walkDeps(basedir, myName)
	for _, name := range deps {
		found := findPkg(name, basedir)
		if found.PType == ptypeUnknown {
			msg := "glide get " + found.Name
			fmt.Printf("\t%s\t(%s)\n", found.Name, msg)
			continue
		}
		if !core && found.PType == ptypeGoroot {
			continue
		}
		fmt.Print(strings.Repeat("\t", level))
		fmt.Printf("%s   (%s)\n", found.Name, found.Path)
		displayTree(found.Path, found.Name, level+1, core)
	}
}

type ptype int8

const (
	ptypeUnknown ptype = iota
	ptypeLocal
	ptypeVendor
	ptypeGopath
	ptypeGoroot
)

type pinfo struct {
	Name, Path string
	PType      ptype
}

func findPkg(name, cwd string) *pinfo {
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
	for _, r := range strings.Split(os.Getenv("GOPATH"), ":") {
		p = filepath.Join(r, "src", name)
		if fi, err = os.Stat(p); err == nil && (fi.IsDir() || isLink(fi)) {
			info.Path = p
			info.PType = ptypeGopath
			return info
		}
	}

	// Check $GOROOT
	for _, r := range strings.Split(os.Getenv("GOROOT"), ":") {
		p = filepath.Join(r, "src", name)
		if fi, err = os.Stat(p); err == nil && (fi.IsDir() || isLink(fi)) {
			info.Path = p
			info.PType = ptypeGoroot
			return info
		}
	}
	return info
}

func isLink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink == os.ModeSymlink
}

func walkDeps(base, myName string) []string {
	externalDeps := []string{}
	filepath.Walk(base, func(path string, fi os.FileInfo, err error) error {
		if excludeSubtree(path, fi) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		pkg, err := build.ImportDir(path, 0)
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
