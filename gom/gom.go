package gom

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
	"github.com/sdboyer/gps"
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

		pkg, _ := util.NormalizeName(gom.name)

		dep := &cfg.Dependency{
			Name: pkg,
		}

		// Check for a specific revision
		if val, ok := gom.options["commit"]; ok {
			dep.Version = val.(string)
		}
		if val, ok := gom.options["tag"]; ok {
			dep.Version = val.(string)
		}
		if val, ok := gom.options["branch"]; ok {
			dep.Branch = val.(string)
		}

		buf = append(buf, dep)
	}

	return buf, nil
}

// AsMetadataPair attempts to extract manifest and lock data from gom metadata.
func AsMetadataPair(dir string) (gps.Manifest, gps.Lock, error) {
	path := filepath.Join(dir, "Gomfile")
	if _, err := os.Stat(path); err != nil {
		return nil, nil, err
	}

	goms, err := parseGomfile(path)
	if err != nil {
		return nil, nil, err
	}

	var l gps.SimpleLock
	m := gps.SimpleManifest{}

	for _, gom := range goms {
		// Do we need to skip this dependency?
		if val, ok := gom.options["skipdep"]; ok && val.(string) == "true" {
			continue
		}

		// Check for custom cloning command
		if _, ok := gom.options["command"]; ok {
			return nil, nil, errors.New("Glide does not support custom Gomfile commands")
		}

		// Check for groups/environments
		if val, ok := gom.options["group"]; ok {
			groups := toStringSlice(val)
			if !stringsContain(groups, "development") && !stringsContain(groups, "production") {
				// right now we only support development and production
				continue
			}
		}

		pkg, _ := util.NormalizeName(gom.name)

		dep := gps.ProjectConstraint{
			Ident: gps.ProjectIdentifier{
				ProjectRoot: gps.ProjectRoot(pkg),
			},
		}

		// Our order of preference for things to put in the manifest are
		//   - Semver
		//   - Version
		//   - Branch
		//   - Revision

		var v gps.UnpairedVersion
		if val, ok := gom.options["tag"]; ok {
			body := val.(string)
			v = gps.NewVersion(body)
			c, err := gps.NewSemverConstraint(body)
			if err != nil {
				c = gps.NewVersion(body)
			}
			dep.Constraint = c
		} else if val, ok := gom.options["branch"]; ok {
			body := val.(string)
			v = gps.NewBranch(body)
			dep.Constraint = gps.NewBranch(body)
		}

		id := gps.ProjectIdentifier{
			ProjectRoot: gps.ProjectRoot(dir),
		}
		var version gps.Version
		if val, ok := gom.options["commit"]; ok {
			body := val.(string)
			if v != nil {
				version = v.Is(gps.Revision(body))
			} else {
				// As with the other third-party system integrations, we're
				// going to choose not to put revisions into a manifest, even
				// though gom has a lot more information than most and the
				// argument could be made for it.
				dep.Constraint = gps.Any()
				version = gps.Revision(body)
			}
		} else if v != nil {
			// This is kinda uncomfortable - lock w/no immut - but OK
			version = v
		}
		l = append(l, gps.NewLockedProject(id, version, nil))

		// TODO We ignore GOOS, GOARCH for now
	}

	return m, l, nil
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
