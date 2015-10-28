package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/yaml"
)

// LinkPackage creates a symlink to the project within the GOPATH.
func LinkPackage(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := c.Get("cfg", "").(*yaml.Config)
	pname := p.Get("path", cfg.Name).(string)

	// Per issue #10, this may be nicer to work with in cases where repos are
	// moved.
	//here := "../.."
	depth := strings.Count(pname, "/")
	here := "../.." + strings.Repeat("/..", depth)

	gopath := Gopath()
	if gopath == "" {
		return nil, fmt.Errorf("$GOPATH appears to be unset")
	}
	if len(pname) == 0 {
		return nil, fmt.Errorf("glide.yaml is missing 'package:'")
	}

	base := path.Dir(pname)
	if base != "." {
		dir := fmt.Sprintf("%s/src/%s", gopath, base)
		if err := os.MkdirAll(dir, os.ModeDir|0755); err != nil {
			return nil, fmt.Errorf("Failed to make directory %s: %s", dir, err)
		}
	}

	ldest := fmt.Sprintf("%s/src/%s", gopath, pname)
	if err := os.Symlink(here, ldest); err != nil {
		if os.IsExist(err) {
			Info("Link to %s already exists. Skipping.\n", ldest)
		} else {
			return nil, fmt.Errorf("Failed to create symlink from %s to %s: %s", gopath, ldest, err)
		}
	}

	return ldest, nil
}
