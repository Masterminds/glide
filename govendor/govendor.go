package govendor

import (
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
)

// Has returns true if this dir has a govendor-flavored vendorfile.
func Has(dir string) bool {
	path := filepath.Join(dir, "vendor/"+Name)
	_, err := os.Stat(path)
	return err == nil
}

func Parse(dir string) ([]*cfg.Dependency, error) {
	path := filepath.Join(dir, "vendor/"+Name)
	if fi, err := os.Stat(path); err != nil || fi.IsDir() {
		return []*cfg.Dependency{}, nil
	}

	msg.Info("Found govendor vendor.json file in %s", gpath.StripBasepath(dir))
	msg.Info("--> Parsing govendor metadata...")
	buf := []*cfg.Dependency{}
	file, err := os.Open(path)
	if err != nil {
		return buf, err
	}
	defer file.Close()

	f := File{}

	if err := f.Unmarshal(file); err != nil {
		return buf, err
	}

	seen := map[string]bool{}
	for _, p := range f.Package {
		pkg, sub := util.NormalizeName(p.PathOrigin())
		msg.Debug("Parsing %s -> %s", pkg, sub)
		processPackage(&seen, &buf, p, pkg, sub)
	}
	return buf, nil
}

func processPackage(seen *map[string]bool, buf *[]*cfg.Dependency, p *Package, pkg, sub string) {
	m := *seen
	if _, ok := m[pkg]; ok {
		if len(sub) == 0 {
			return
		}

		for _, dep := range *buf {
			if dep.Name == pkg {
				subRef := getReferenceFromPackage(p)
				if dep.Reference != subRef {
					// Add conflicting subpackage as 1st class package
					// and let the user resolve it.
					// First `glide update` will report these.
					processPackage(seen, buf, p, p.PathOrigin(), "")
					return
				}
				dep.Subpackages = append(dep.Subpackages, sub)
			}
		}
	} else {
		m[pkg] = true
		seen = &m

		dep := &cfg.Dependency{
			Name:      pkg,
			Reference: getReferenceFromPackage(p),
		}

		msg.Info("Parsed %s (%s)", pkg, dep.Reference)

		if len(sub) > 0 {
			dep.Subpackages = []string{sub}
		}
		*buf = append(*buf, dep)
	}
	return
}

func getReferenceFromPackage(pkg *Package) string {
	// TODO: Prefer Version over VersionExact?
	if pkg.VersionExact != "" {
		return pkg.VersionExact
	}

	if pkg.Version != "" {
		return pkg.Version
	}

	return pkg.Revision
}
