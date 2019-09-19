// Package trash reads Trash's vendor files.
//
// It is not a complete implementaton of Trash.
package trash

import (
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/trash/conf"
)

// Has indicates whether a Trash file exists.
func Has(dir string) bool {
	path := filepath.Join(dir, "vendor.conf")
	_, err := os.Stat(path)
	return err == nil
}

// Parse parses a Trash vendor.conf file.
func Parse(dir string) ([]*cfg.Dependency, error) {
	path := filepath.Join(dir, "vendor.conf")
	if i, err := os.Stat(path); err != nil {
		return []*cfg.Dependency{}, nil
	} else if i.IsDir() {
		msg.Info("vendor.conf is a directory.\n")
		return []*cfg.Dependency{}, nil
	}
	msg.Info("Found vendor.conf file in %s", gpath.StripBasepath(dir))
	msg.Info("--> Parsing Trash metadata...")

	buf := []*cfg.Dependency{}

	trashconf, err := conf.Parse(path)
	if err != nil {
		return buf, err
	}

	for _, trashimport := range trashconf.Imports {
		dep := &cfg.Dependency{}
		dep.Name = trashimport.Package
		dep.Reference = trashimport.Version
		dep.Repository = trashimport.Repo

		buf = append(buf, dep)
	}

	return buf, nil
}
