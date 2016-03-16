package path

import (
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/msg"
)

// StripVcs removes VCS metadata (.git, .hg, .bzr, .svn) from the vendor/
// directory.
func StripVcs() error {
	if _, err := os.Stat(VendorDir); err != nil {
		if os.IsNotExist(err) {
			msg.Debug("Vendor directory does not exist.")
		}

		return err
	}
	return filepath.Walk(VendorDir, stripHandler)
}

func stripHandler(path string, info os.FileInfo, err error) error {

	name := info.Name()
	if name == ".git" || name == ".bzr" || name == ".svn" || name == ".hg" {
		if _, err := os.Stat(path); err == nil {
			if info.IsDir() {
				msg.Info("Removing: %s", path)
				return os.RemoveAll(path)
			}

			msg.Debug("%s is not a directory. Skipping removal", path)
			return nil
		}
	}
	return nil
}
