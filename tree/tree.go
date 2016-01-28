package tree

import (
	"container/list"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
)

// Display displays a tree view of the given project.
//
// FIXME: The output formatting could use some TLC.
func Display(b *util.BuildCtxt, basedir, myName string, level int, core bool, l *list.List) {
	deps := walkDeps(b, basedir, myName)
	for _, name := range deps {
		found := findPkg(b, name, basedir)
		if found.Loc == dependency.LocUnknown {
			m := "glide get " + found.Name
			msg.Puts("\t%s\t(%s)", found.Name, m)
			continue
		}
		if !core && found.Loc == dependency.LocGoroot || found.Loc == dependency.LocCgo {
			continue
		}
		msg.Print(strings.Repeat("|\t", level-1) + "|-- ")

		f := findInList(found.Name, l)
		if f == true {
			msg.Puts("(Recursion) %s   (%s)", found.Name, found.Path)
		} else {
			// Every branch in the tree is a copy to handle all the branches
			cl := copyList(l)
			cl.PushBack(found.Name)
			msg.Puts("%s   (%s)", found.Name, found.Path)
			Display(b, found.Path, found.Name, level+1, core, cl)
		}
	}
}

func walkDeps(b *util.BuildCtxt, base, myName string) []string {
	externalDeps := []string{}
	filepath.Walk(base, func(path string, fi os.FileInfo, err error) error {
		if !dependency.IsSrcDir(fi) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		pkg, err := b.ImportDir(path, 0)
		if err != nil {
			if !strings.HasPrefix(err.Error(), "no buildable Go source") {
				msg.Warn("Error: %s (%s)", err, path)
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

func findPkg(b *util.BuildCtxt, name, cwd string) *dependency.PkgInfo {
	var fi os.FileInfo
	var err error
	var p string

	info := &dependency.PkgInfo{
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
		// Previously there was a check on the loop that wd := "/". The path
		// "/" is a POSIX path so this fails on Windows. Now the check is to
		// make sure the same wd isn't seen twice. When the same wd happens
		// more than once it's the beginning of looping on the same location
		// which is the top level.
		pwd := ""
		for wd := abs; wd != pwd; wd = filepath.Dir(wd) {
			pwd = wd

			// Don't look for packages outside the GOPATH
			// Note, the GOPATH may or may not end with the path separator.
			// The output of filepath.Dir does not the the path separator on the
			// end so we need to test both.
			if wd == b.GOPATH || wd+string(os.PathSeparator) == b.GOPATH {
				break
			}
			p = filepath.Join(wd, "vendor", name)
			if fi, err = os.Stat(p); err == nil && (fi.IsDir() || gpath.IsLink(fi)) {
				info.Path = p
				info.Loc = dependency.LocVendor
				info.Vendored = true
				return info
			}
		}
	}
	// Check $GOPATH
	for _, r := range strings.Split(b.GOPATH, ":") {
		p = filepath.Join(r, "src", name)
		if fi, err = os.Stat(p); err == nil && (fi.IsDir() || gpath.IsLink(fi)) {
			info.Path = p
			info.Loc = dependency.LocGopath
			return info
		}
	}

	// Check $GOROOT
	for _, r := range strings.Split(b.GOROOT, ":") {
		p = filepath.Join(r, "src", name)
		if fi, err = os.Stat(p); err == nil && (fi.IsDir() || gpath.IsLink(fi)) {
			info.Path = p
			info.Loc = dependency.LocGoroot
			return info
		}
	}

	// Finally, if this is "C", we're dealing with cgo
	if name == "C" {
		info.Loc = dependency.LocCgo
	}

	return info
}

// copyList copies an existing list to a new list.
func copyList(l *list.List) *list.List {
	n := list.New()
	for e := l.Front(); e != nil; e = e.Next() {
		n.PushBack(e.Value.(string))
	}
	return n
}

// findInList searches a list haystack for a string needle.
func findInList(n string, l *list.List) bool {
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value.(string) == n {
			return true
		}
	}

	return false
}
