package cmd

import (
	"os"
	"path"

	"github.com/Masterminds/cookoo"
)

// Recurse does glide installs on dependent packages.
// Recurse looks in all known packages for a glide.yaml files and installs for
// each one it finds.
func Recurse(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	Info("Checking dependencies for updates.\n")
	conf := p.Get("conf", &Config{}).(*Config)
	vend, _ := VendorPath(c)

	if len(conf.Imports) == 0 {
		Info("No imports.\n")
	}

	// Look in each package to see whether it has a glide.yaml, and no vendor/
	for _, imp := range conf.Imports {
		Info("Looking in %s for a glide.yaml file.\n", imp.Name)
		if needsGlideUp(path.Join(vend, imp.Name)) {
			Info("Package %s needs `glide up`\n", imp.Name)
			// How do we want to do this? Should we run the glide command,
			// which would allow environmental control, or should we just
			// run the update route in that directory?
		}
	}

	// Run `glide up`
	return nil, nil
}

func needsGlideUp(dir string) bool {
	stat, err := os.Stat(path.Join(dir, "glide.yaml"))
	if err != nil || stat.IsDir() {
		return false
	}

	// Should probably see if vendor is there and non-empty.

	return true
}
