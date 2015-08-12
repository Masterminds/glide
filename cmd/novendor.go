package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/cookoo"
)

// NoVendor takes a path and returns all subpaths that are not vendor directories.
//
// It is not recursive.
//
// If the given path is a file, it returns that path unaltered.
//
// If the given path is a directory, it scans all of the immediate children,
// and returns all of the go files and directories that are not vendor.
func NoVendor(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	path := p.Get("path", ".").(string)

	return noVend(path)
}

// Take a list of paths and print a single string with space-separated paths.
func PathString(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	paths := p.Get("paths", []string{}).([]string)
	s := strings.Join(paths, " ")
	fmt.Println(s)
	return nil, nil
}

// noVend takes a directory and returns a list of Go-like files or directories,
// provided the directory is not a vendor directory.
func noVend(path string) ([]string, error) {

	info, err := os.Stat(path)
	if err != nil {
		return []string{}, err
	}

	if !info.IsDir() {
		return []string{path}, nil
	}

	res := []string{}
	f, err := os.Open(path)
	if err != nil {
		return res, err
	}

	fis, err := f.Readdir(0)
	if err != nil {
		return res, err
	}

	for _, fi := range fis {
		full := filepath.Join(path, fi.Name())
		if fi.IsDir() && !isVend(fi) {
			res = append(res, full)
		} else if !fi.IsDir() && isGoish(fi) {
			res = append(res, full)
		}
	}
	return res, nil
}

func isVend(fi os.FileInfo) bool {
	return fi.Name() == "vendor"
}

func isGoish(fi os.FileInfo) bool {
	return filepath.Ext(fi.Name()) == ".go"
}
