package action

import (
	"fmt"
	"os"
	"path/filepath"
)

const DefaultGlideFile = "glide.yaml"

// VendorDir is the name of the directory that holds vendored dependencies.
//
// As of Go 1.5, this is always vendor.
var VendorDir = "vendor"

// GlideFile is the name of the Glide file.
//
// Setting this is not concurrency safe. For consistency, it should really
// only be set once, at startup, or not at all.
var GlideFile = DefaultGlideFile

// VendorPath calculates the path to the vendor directory.
//
// Based on working directory, VendorDir and GlideFile, this attempts to
// guess the location of the vendor directory.
func VendorPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Find the directory that contains glide.yaml
	yamldir, err := GlideWD(cwd)
	if err != nil {
		return cwd, err
	}

	gopath := filepath.Join(yamldir, VendorDir)

	return gopath, nil
}

// GlideWD finds the working directory of the glide.yaml file, starting at dir.
//
// If the glide file is not found in the current directory, it recurses up
// a directory.
func GlideWD(dir string) (string, error) {
	fullpath := filepath.Join(dir, GlideFile)

	if _, err := os.Stat(fullpath); err == nil {
		return dir, nil
	}

	base := filepath.Dir(dir)
	if base == dir {
		return "", fmt.Errorf("Cannot resolve parent of %s", base)
	}

	return GlideWD(base)
}
