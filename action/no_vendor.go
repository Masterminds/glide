package action

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/msg"
)

// NoVendor generates a list of source code directories, excepting `vendor/`.
//
// If "onlyGo" is true, only folders that have Go code in them will be returned.
//
// If suffix is true, this will append `/...` to every directory.
func NoVendor(path string, onlyGo, suffix bool) {
	// This is responsible for printing the results of noVend.
	paths, err := noVend(path, onlyGo, suffix)
	if err != nil {
		msg.Err("Failed to walk file tree: %s", err)
		msg.Warn("FIXME: NoVendor should exit with non-zero exit code.")
		return
	}

	for _, p := range paths {
		msg.Puts(p)
	}
}

// noVend takes a directory and returns a list of Go-like files or directories,
// provided the directory is not a vendor directory.
//
// If onlyGo is true, this will filter out all directories that do not contain
// ".go" files.
//
// TODO: Should we move this to its own package?
func noVend(path string, onlyGo, suffix bool) ([]string, error) {

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
		res = hasGoSource(res, suffix)
	}

	if cur {
		res = append(res, ".")
	}

	return res, nil
}

// hasGoSource returns a list of directories that contain Go source.
func hasGoSource(dirs []string, suffix bool) []string {
	suf := "/"
	if suffix {
		suf = "/..."
	}
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
			buf = append(buf, "./"+d+suf)
		}
	}
	return buf
}

// isVend returns true of this directory is a vendor directory.
//
// TODO: Should we return true for Godeps directory?
func isVend(fi os.FileInfo) bool {
	return fi.Name() == "vendor"
}

// exclude returns true if the directory should be excluded by Go toolchain tools.
//
// Examples: directories prefixed with '.' or '_'.
func exclude(fi os.FileInfo) bool {
	if strings.HasPrefix(fi.Name(), "_") {
		return true
	}
	if strings.HasPrefix(fi.Name(), ".") {
		return true
	}
	return false
}

// isGoish returns true if the file appears to be Go source.
func isGoish(fi os.FileInfo) bool {
	return filepath.Ext(fi.Name()) == ".go"
}
