package gom

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
)

// Has returns true if this dir has a Gomfile.
func Has(dir string) bool {
	path := filepath.Join(dir, "Gomfile")
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

// Parse parses a Gomfile.
func Parse(dir string) ([]*cfg.Dependency, error) {
	path := filepath.Join(dir, "Gomfile")
	if fi, err := os.Stat(path); err != nil || fi.IsDir() {
		return []*cfg.Dependency{}, nil
	}

	msg.Info("Found Gomfile in %s", gpath.StripBasepath(dir))
	msg.Info("--> Parsing Gomfile metadata...")
	buf := []*cfg.Dependency{}

	goms, err := parseGomfile(path)
	if err != nil {
		return []*cfg.Dependency{}, err
	}

	for _, gom := range goms {
		// Do we need to skip this dependency?
		if val, ok := gom.options["skipdep"]; ok && val.(string) == "true" {
			continue
		}

		// Check for custom cloning command
		if _, ok := gom.options["command"]; ok {
			return []*cfg.Dependency{}, errors.New("Glide does not support custom Gomfile commands")
		}

		// Check for groups/environments
		if val, ok := gom.options["group"]; ok {
			groups := toStringSlice(val)
			if !stringsContain(groups, "development") && !stringsContain(groups, "production") {
				// right now we only support development and production
				msg.Info("Skipping dependency '%s' because it isn't in the development or production group", gom.name)
				continue
			}
		}

		pkg, sub := util.NormalizeName(gom.name)

		dep := &cfg.Dependency{
			Name: pkg,
		}

		if len(sub) > 0 {
			dep.Subpackages = []string{sub}
		}

		// Check for a specific revision
		if val, ok := gom.options["commit"]; ok {
			dep.Reference = val.(string)
		}
		if val, ok := gom.options["tag"]; ok {
			dep.Reference = val.(string)
		}
		if val, ok := gom.options["branch"]; ok {
			dep.Reference = val.(string)
		}

		// Parse goos and goarch
		if val, ok := gom.options["goos"]; ok {
			dep.Os = toStringSlice(val)
		}
		if val, ok := gom.options["goarch"]; ok {
			dep.Arch = toStringSlice(val)
		}

		buf = append(buf, dep)
	}

	return buf, nil
}

func stringsContain(v []string, key string) bool {
	for _, s := range v {
		if s == key {
			return true
		}
	}
	return false
}

func toStringSlice(v interface{}) []string {
	if v, ok := v.(string); ok {
		return []string{v}
	}

	if v, ok := v.([]string); ok {
		return v
	}

	return []string{}
}
