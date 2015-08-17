package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/cookoo"
)

// ReadyToGlide fails if the environment is not sufficient for using glide.
//
// Most importantly, it fails if glide.yaml is not present in the current
// working directory.
func ReadyToGlide(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fname := p.Get("filename", "glide.yaml").(string)
	if _, err := os.Stat(fname); err != nil {
		cwd, _ := os.Getwd()
		return false, fmt.Errorf("%s is missing from %s", fname, cwd)
	}
	return true, nil
}

// Return the path to the vendor directory.
func VendorPath(c cookoo.Context) (string, error) {
	vendor := c.Get("VendorDir", "vendor").(string)
	filename := c.Get("yaml", "glide.yaml").(string)
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Find the directory that contains glide.yaml
	yamldir, err := glideWD(cwd, filename)
	if err != nil {
		return cwd, err
	}

	gopath := filepath.Join(yamldir, vendor)

	return gopath, nil
}

func glideWD(dir, filename string) (string, error) {
	fullpath := filepath.Join(dir, filename)

	if _, err := os.Stat(fullpath); err == nil {
		return dir, nil
	}

	base := filepath.Dir(dir)
	if base == dir {
		return "", fmt.Errorf("Cannot resolve parent of %s", base)
	}

	return glideWD(base, filename)
}
