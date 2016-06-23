package gom

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
	"github.com/sdboyer/vsolver"
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

// AsMetadataPair attempts to extract manifest and lock data from gom metadata.
func AsMetadataPair(dir string) (vsolver.Manifest, vsolver.Lock, error) {
	path := filepath.Join(dir, "Gomfile")
	if _, err := os.Stat(path); err != nil {
		return nil, nil, err
	}

	goms, err := parseGomfile(path)
	if err != nil {
		return nil, nil, err
	}

	var l vsolver.SimpleLock
	m := vsolver.SimpleManifest{
		N: vsolver.ProjectName(dir),
	}

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

		dep := vsolver.ProjectDep{
			Ident: vsolver.ProjectIdentifier{
				LocalName: vsolver.ProjectName(pkg),
			},
		}

		// Our order of preference for things to put in the manifest are
		//   - Semver
		//   - Version
		//   - Branch
		//   - Revision

		var v vsolver.UnpairedVersion
		if val, ok := gom.options["tag"]; ok {
			body := val.(string)
			v = vsolver.NewVersion(body)
			c, err := vsolver.NewSemverConstraint(body)
			if err != nil {
				c = vsolver.NewVersion(body)
			}
			dep.Constraint = c
		} else if val, ok := gom.options["branch"]; ok {
			body := val.(string)
			v = vsolver.NewBranch(body)
			dep.Constraint = vsolver.NewBranch(body)
		}

		if val, ok := gom.options["commit"]; ok {
			body := val.(string)
			if v != nil {
				v.Is(vsolver.Revision(body))
				l = append(l, vsolver.NewLockedProject(vsolver.ProjectName(dir), v, dir, dir, nil))
			} else {
				// As with the other third-party system integrations, we're
				// going to choose not to put revisions into a manifest, even
				// though gom has a lot more information than most and the
				// argument could be made for it.
				dep.Constraint = vsolver.Any()
				l = append(l, vsolver.NewLockedProject(vsolver.ProjectName(dir), vsolver.Revision(body), dir, dir, nil))
			}
		} else if v != nil {
			// This is kinda uncomfortable - lock w/no immut - but OK
			l = append(l, vsolver.NewLockedProject(vsolver.ProjectName(dir), v, dir, dir, nil))
		}

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
