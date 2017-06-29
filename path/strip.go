package path

import (
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/godep/strip"
	"github.com/Masterminds/glide/msg"
)

// StripVendor removes nested vendor and Godeps/_workspace/ directories.
func StripVendor() error {
	searchPath, _ := Vendor()
	if _, err := os.Stat(searchPath); err != nil {
		if os.IsNotExist(err) {
			msg.Debug("Vendor directory does not exist.")
		}

		return err
	}

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		// Skip the base vendor directory
		if path == searchPath {
			return nil
		}

		name := info.Name()
		if name == "vendor" {
			if _, err := os.Stat(path); err == nil {
				if info.IsDir() {
					msg.Info("Removing: %s", path)
					return CustomRemoveAll(path)
				}

				msg.Debug("%s is not a directory. Skipping removal", path)
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return strip.GodepWorkspace(searchPath)
}
