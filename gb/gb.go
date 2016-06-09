package gb

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
)

// Has returns true if this dir has a GB-flavored manifest file.
func Has(dir string) bool {
	path := filepath.Join(dir, "vendor/manifest")
	_, err := os.Stat(path)
	return err == nil
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
		pkg, sub := util.NormalizeName(d.Importpath)
		if _, ok := seen[pkg]; ok {
			if len(sub) == 0 {
				continue
			}
			for _, dep := range buf {
				if dep.Name == pkg {
					dep.Subpackages = append(dep.Subpackages, sub)
				}
			}
		} else {
			seen[pkg] = true
			dep := &cfg.Dependency{
				Name:       pkg,
				Reference:  d.Revision,
				Repository: d.Repository,
			}
			if len(sub) > 0 {
				dep.Subpackages = []string{sub}
			}
			buf = append(buf, dep)
		}
	}
	return buf, nil
}
