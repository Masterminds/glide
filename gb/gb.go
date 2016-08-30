package gb

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
	"github.com/sdboyer/gps"
)

// Has returns true if this dir has a GB-flavored manifest file.
func Has(dir string) bool {
	path := filepath.Join(dir, "vendor/manifest")
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

// Parse parses a GB-flavored manifest file.
func Parse(dir string) ([]*cfg.Dependency, error) {
	path := filepath.Join(dir, "vendor/manifest")
	if fi, err := os.Stat(path); err != nil || fi.IsDir() {
		return []*cfg.Dependency{}, nil
	}

	msg.Info("Found GB manifest file in %s", gpath.StripBasepath(dir))
	msg.Info("--> Parsing GB metadata...")
	buf := []*cfg.Dependency{}
	file, err := os.Open(path)
	if err != nil {
		return buf, err
	}
	defer file.Close()

	man := Manifest{}

	dec := json.NewDecoder(file)
	if err := dec.Decode(&man); err != nil {
		return buf, err
	}

	seen := map[string]bool{}

	for _, d := range man.Dependencies {
		// TODO(sdboyer) move to the corresponding SourceManager call...though
		// that matters less once gps caches these results
		pkg, _ := util.NormalizeName(d.Importpath)
		if !seen[pkg] {
			seen[pkg] = true
			dep := &cfg.Dependency{
				Name:       pkg,
				Constraint: cfg.DeduceConstraint(d.Revision),
				Repository: d.Repository,
			}
			buf = append(buf, dep)
		}
	}
	return buf, nil
}

// AsMetadataPair attempts to extract manifest and lock data from gb metadata.
func AsMetadataPair(dir string) (m []*cfg.Dependency, l *cfg.Lockfile, err error) {
	path := filepath.Join(dir, "vendor/manifest")
	if _, err = os.Stat(path); err != nil {
		return
	}

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	man := Manifest{}

	dec := json.NewDecoder(file)
	if err = dec.Decode(&man); err != nil {
		return
	}

	seen := map[string]bool{}

	for _, d := range man.Dependencies {
		pkg, _ := util.NormalizeName(d.Importpath)
		if !seen[pkg] {
			seen[pkg] = true
			dep := &cfg.Dependency{
				Name:       pkg,
				Constraint: gps.Any(),
				Repository: d.Repository,
			}
			m = append(m, dep)
			l.Imports = append(l.Imports, &cfg.Lock{Name: pkg, Revision: d.Revision})
		}
	}
	return
}
