package dependency

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

// DeleteUnused removes packages from vendor/ that are no longer used.
//
// TODO: This should work off of a Lock file, not glide.yaml.
func DeleteUnused(conf *cfg.Config) error {
	vpath, err := gpath.Vendor()
	if err != nil {
		return err
	}
	if vpath == "" {
		return errors.New("Vendor not set")
	}

	// Build directory tree of what to keep.
	var pkgList []string
	for _, dep := range conf.Imports {
		pkgList = append(pkgList, dep.Name)
	}

	var searchPath string
	var markForDelete []string
	// Callback function for filepath.Walk to delete packages not in yaml file.
	fn := func(path string, info os.FileInfo, err error) error {
		// Bubble up the error
		if err != nil {
			return err
		}

		if info.IsDir() == false || path == searchPath || path == vpath {
			return nil
		}

		localPath := strings.TrimPrefix(path, searchPath)
		keep := false

		// First check if the path has a prefix that's a specific package. If
		// so we keep it to keep the package.
		for _, name := range pkgList {
			if strings.HasPrefix(localPath, name) {
				keep = true
			}
		}

		// If a package is, for example, github.com/Masterminds/glide the
		// previous look will not mark the directories github.com or
		// github.com/Masterminds to keep. Here we see if these names prefix
		// and packages we know about to mark as keepers.
		if keep == false {
			for _, name := range pkgList {
				if strings.HasPrefix(name, localPath) {
					keep = true
				}
			}
		}

		// If the parent directory has already been marked for delete this
		// directory doesn't need to be marked.
		for _, markedDirectory := range markForDelete {
			if strings.HasPrefix(path, markedDirectory) {
				return nil
			}
		}

		// Remove the directory if we are not keeping it.
		if keep == false {
			// Mark for deletion
			markForDelete = append(markForDelete, path)
		}

		return nil
	}

	// Walk vendor directory
	searchPath = vpath + string(os.PathSeparator)
	err = filepath.Walk(searchPath, fn)
	if err != nil {
		return err
	}

	// Perform the actual delete.
	for _, path := range markForDelete {
		localPath := strings.TrimPrefix(path, searchPath)
		msg.Info("Removing unused package: %s\n", localPath)
		rerr := os.RemoveAll(path)
		if rerr != nil {
			return rerr
		}
	}

	return nil
}
