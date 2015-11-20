package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/yaml"
)

// DeleteUnusedPackages removes packages from vendor/ that are no longer used.
func DeleteUnusedPackages(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	// Conditional opt-in to removed unused dependencies.
	optIn := p.Get("optIn", false).(bool)
	if optIn != true {
		return nil, nil
	}

	vpath, err := VendorPath(c)
	if err != nil {
		return nil, err
	}
	if vpath == "" {
		return false, errors.New("Vendor not set")
	}

	// Build directory tree of what to keep.
	cfg := p.Get("conf", nil).(*yaml.Config)
	var pkgList []string
	for _, dep := range cfg.Imports {
		for _, sub := range dep.Subpackages {
			pkgList = append(pkgList, dep.Name+"/"+sub)
		}
		if len(dep.Subpackages) == 0 {
			pkgList = append(pkgList, dep.Name)
		}
	}

	// Callback function for filepath.Walk to delete packages not in yaml file.
	var searchPath string

	var markForDelete []string

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
			if localPath == name || strings.HasPrefix(localPath, name+"/") || strings.HasSuffix(name, "/.") && localPath == name[:len(name)-2] {
				keep = true
				break
			}
		}

		// If a package is, for example, github.com/Masterminds/glide the
		// previous look will not mark the directories github.com or
		// github.com/Masterminds to keep. Here we see if these names prefix
		// and packages we know about to mark as keepers.
		if keep == false {
			for _, name := range pkgList {
				if strings.HasPrefix(name, localPath+"/") {
					keep = true
					break
				}
			}
		}

		// If the parent directory has already been marked for delete this
		// directory doesn't need to be marked.
		for _, markedDirectory := range markForDelete {
			if strings.HasPrefix(path, markedDirectory+"/") {
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
	searchPath = vpath + "/"
	err = filepath.Walk(searchPath, fn)
	if err != nil {
		return false, err
	}

	// Perform the actual delete.
	for _, path := range markForDelete {
		localPath := strings.TrimPrefix(path, searchPath)
		Info("Removing unused package: %s\n", localPath)
		rerr := os.RemoveAll(path)
		if rerr != nil {
			return false, rerr
		}
	}

	return nil, nil
}
