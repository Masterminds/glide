package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/cookoo"
)

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
