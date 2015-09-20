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
	gonly := p.Get("onlyGo", true).(bool)

	return noVend(path, gonly)
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
//
// If onlyGo is true, this will filter out all directories that do not contain
// ".go" files.
func noVend(path string, onlyGo bool) ([]string, error) {

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

	cur := false

	for _, fi := range fis {
		if exclude(fi) {
			continue
		}

		full := filepath.Join(path, fi.Name())
		if fi.IsDir() && !isVend(fi) {
			p := "./" + full + "/..."
			res = append(res, p)
		} else if !fi.IsDir() && isGoish(fi) {
			//res = append(res, full)
			cur = true
		}
	}

	// Filter out directories that do not contain Go code
	if onlyGo {
		res = hasGoSource(res)
	}

	if cur {
		res = append(res, ".")
	}

	return res, nil
}

func hasGoSource(dirs []string) []string {
	buf := []string{}
	for _, d := range dirs {
		d := filepath.Dir(d)
		found := false
		walker := func(p string, fi os.FileInfo, err error) error {
			// Dumb optimization
			if found {
				return nil
			}

			// If the file ends with .go, report a match.
			if strings.ToLower(filepath.Ext(p)) == ".go" {
				found = true
			}

			return nil
		}
		filepath.Walk(d, walker)

		if found {
			buf = append(buf, d)
		}
	}
	return buf
}

func isVend(fi os.FileInfo) bool {
	return fi.Name() == "vendor"
}

func exclude(fi os.FileInfo) bool {
	if strings.HasPrefix(fi.Name(), "_") {
		return true
	}
	if strings.HasPrefix(fi.Name(), ".") {
		return true
	}
	return false
}

func isGoish(fi os.FileInfo) bool {
	return filepath.Ext(fi.Name()) == ".go"
}
