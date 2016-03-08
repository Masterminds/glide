// Package importer imports dependency configuration from Glide, Godep, GPM, GB and gom
package importer

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/gb"
	"github.com/Masterminds/glide/godep"
	"github.com/Masterminds/glide/gom"
	"github.com/Masterminds/glide/gpm"
)

var i = &DefaultImporter{}

// Import uses the DefaultImporter to import from Glide, Godep, GPM, GB and gom.
func Import(path string) (bool, []*cfg.Dependency, error) {
	return i.Import(path)
}

// Importer enables importing depenency configuration.
type Importer interface {

	// Import imports dependency configuration. It returns:
	// - A bool if any configuration was found.
	// - []*cfg.Dependency containing dependency configuration if any is found.
	// - An error if one was reported.
	Import(path string) (bool, []*cfg.Dependency, error)
}

// DefaultImporter imports from Glide, Godep, GPM, GB and gom.
type DefaultImporter struct{}

// Import tries to import configuration from Glide, Godep, GPM, GB and gom.
func (d *DefaultImporter) Import(path string) (bool, []*cfg.Dependency, error) {

	// Try importing from Glide first.
	p := filepath.Join(path, "glide.yaml")
	if _, err := os.Stat(p); err == nil {
		// We found glide configuration.
		yml, err := ioutil.ReadFile(p)
		if err != nil {
			return false, []*cfg.Dependency{}, err
		}
		conf, err := cfg.ConfigFromYaml(yml)
		if err != nil {
			return false, []*cfg.Dependency{}, err
		}
		return true, conf.Imports, nil
	}

	// Try importing from Godep
	if godep.Has(path) {
		deps, err := godep.Parse(path)
		if err != nil {
			return false, []*cfg.Dependency{}, err
		}
		return true, deps, nil
	}

	// Try importing from GPM
	if gpm.Has(path) {
		deps, err := gpm.Parse(path)
		if err != nil {
			return false, []*cfg.Dependency{}, err
		}
		return true, deps, nil
	}

	// Try importin from GB
	if gb.Has(path) {
		deps, err := gb.Parse(path)
		if err != nil {
			return false, []*cfg.Dependency{}, err
		}
		return true, deps, nil
	}

	// Try importing from gom
	if gom.Has(path) {
		deps, err := gom.Parse(path)
		if err != nil {
			return false, []*cfg.Dependency{}, err
		}
		return true, deps, nil
	}

	// When none are found.
	return false, []*cfg.Dependency{}, nil
}
