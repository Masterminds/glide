package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/gb"
)

func HasGbManifest(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	dir := cookoo.GetString("dir", "", p)
	path := filepath.Join(dir, "vendor/manifest")
	_, err := os.Stat(path)
	return err == nil, nil
}

// GbManifest
//
// Params:
// 	- dir (string): The directory where the manifest file is located.
// Returns:
//
func GbManifest(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	dir := cookoo.GetString("dir", ".", p)
	return parseGbManifest(dir)
}

func parseGbManifest(dir string) ([]*Dependency, error) {
	path := filepath.Join(dir, "vendor/manifest")
	if fi, err := os.Stat(path); err != nil || fi.IsDir() {
		return []*Dependency{}, nil
	}

	Info("Found GB manifest file.\n")
	buf := []*Dependency{}
	file, err := os.Open(path)
	if err != nil {
		return buf, err
	}
	defer file.Close()

	man := gb.Manifest{}

	dec := json.NewDecoder(file)
	if err := dec.Decode(&man); err != nil {
		return buf, err
	}

	seen := map[string]bool{}

	for _, d := range man.Dependencies {
		pkg, sub := NormalizeName(d.Importpath)
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
			dep := &Dependency{
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
