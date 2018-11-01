package path

import (
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/godep/strip"
	"github.com/Masterminds/glide/msg"
)

func getWalkFunction(searchPath string, removeAll func(p string) error) func(path string,
	info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the base vendor directory
		if path == searchPath {
			return nil
		}
		// Skip if FileInfo is nil
		if info == nil {
			return nil
		}

		if info.Name() == "vendor" && info.IsDir() {
			msg.Info("Removing: %s", path)
			err = removeAll(path)
			if nil != err {
				return err
			}
			return filepath.SkipDir
		}
		return nil
	}
}

// StripVendor removes nested vendor and Godeps/_workspace/ directories.
func StripVendor() error {
	searchPath, _ := Vendor()
	if _, err := os.Stat(searchPath); err != nil {
		if os.IsNotExist(err) {
			msg.Debug("Vendor directory does not exist.")
		}

		return err
	}

	err := filepath.Walk(searchPath, getWalkFunction(searchPath, CustomRemoveAll))

	if err != nil {
		return err
	}

	return strip.GodepWorkspace(searchPath)
}
