package repo

import (
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

// VendoredCleanup cleans up vendored codebases after an update.
//
// This should _only_ be run for installations that do not want VCS repos inside
// of the vendor/ directory.
func VendoredCleanup(conf *cfg.Config) error {
	vend, err := gpath.Vendor()
	if err != nil {
		return err
	}

	for _, dep := range conf.Imports {
		if dep.UpdateAsVendored == true {
			msg.Info("Cleaning up vendored package %s\n", dep.Name)

			// Remove the VCS directory
			cwd := filepath.Join(vend, dep.Name)
			repo, err := dep.GetRepo(cwd)
			if err != nil {
				msg.Err("Error cleaning up %s:%s", dep.Name, err)
				continue
			}
			t := repo.Vcs()
			err = os.RemoveAll(cwd + string(os.PathSeparator) + "." + string(t))
			if err != nil {
				msg.Err("Error cleaning up VCS dir for %s:%s", dep.Name, err)
			}
		}

	}

	return nil
}
