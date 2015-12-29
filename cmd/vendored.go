package cmd

import (
	"os"
	"path/filepath"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/cfg"
)

// If we are updating the vendored dependencies. That is those stored in the
// local project VCS.
var updatingVendored = false

// VendoredSetup is a command that does the setup for vendored directories.
// If enabled (via update) it marks vendored directories that are being updated
// and removed the old code. This should be a prefix to UpdateImports and
// VendoredCleanUp should be a suffix to UpdateImports.
func VendoredSetup(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	update := p.Get("update", false).(bool)
	conf := p.Get("conf", nil).(*cfg.Config)

	updatingVendored = update

	return conf, nil
}

// VendoredCleanUp is a command that cleans up vendored codebases after an update.
// If enabled (via update) it removes the VCS info from updated vendored
// packages. This should be a suffix to UpdateImports and  VendoredSetup should
// be a prefix to UpdateImports.
func VendoredCleanUp(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	update := p.Get("update", true).(bool)
	if update != true {
		return false, nil
	}
	conf := p.Get("conf", nil).(*cfg.Config)

	vend, err := VendorPath(c)
	if err != nil {
		return false, err
	}

	for _, dep := range conf.Imports {
		if dep.UpdateAsVendored == true {
			Info("Cleaning up vendored package %s\n", dep.Name)

			// Remove the VCS directory
			cwd := filepath.Join(vend, filepath.FromSlash(dep.Name))
			repo, err := dep.GetRepo(cwd)
			if err != nil {
				Error("Error cleaning up %s:%s", dep.Name, err)
				continue
			}
			t := repo.Vcs()
			err = os.RemoveAll(cwd + string(os.PathSeparator) + "." + string(t))
			if err != nil {
				Error("Error cleaning up VCS dir for %s:%s", dep.Name, err)
			}
		}

	}

	return true, nil
}
