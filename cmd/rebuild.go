package cmd

import (
	"github.com/Masterminds/cookoo"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"fmt"
)

// Rebuild run 'go build' in a directory.
//
// Params:
// 	- conf: the *Config.
//
func Rebuild(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)
	gopath := os.Getenv("GOPATH")

	Info("Building dependencies.\n")

	if len(cfg.Imports) == 0 {
		Info("No dependencies found. Nothing built.\n")
		return true, nil
	}

	for _, dep := range cfg.Imports {
		if err := buildDep(c, dep, gopath); err != nil {
			Warn("Failed to build %s: %s\n", dep.Name, err)
		}
	}

	return true, nil
}

func buildDep(c cookoo.Context, dep *Dependency, gopath string) error {
	if len(dep.Subpackages) == 0 {
		buildPath(c, dep.Name)
	}

	for _, pkg := range dep.Subpackages {

		if pkg == "**" {
			//buildAll(c, path.Join(gopath, "src", dep.Name))
			Info("Building all packages in %s\n", dep.Name)
			buildPath(c, path.Join(dep.Name, "..."))
		} else {
			buildPath(c, path.Join(dep.Name, pkg))
		}
	}

	return nil
}

func joinAndResolv(c cookoo.Context, parts ...string) ([]string, error) {
	path := path.Join(parts...)
	return filepath.Glob(path)
}

func buildPaths(c cookoo.Context, paths []string) error {
	for _, path := range paths {
		if err := buildPath(c, path); err != nil {
			return err
		}
	}

	return nil
}

func buildPath(c cookoo.Context, path string) error {
	/*
	if err := os.Chdir(path); err != nil {
		//return err
		Warn("%s is not a directory. Skipping.\n", path)
	}
	*/

	//out, err := exec.Command("go", "build", "./...").CombinedOutput()
	Info("Running go build %s\n", path)
	out, err := exec.Command("go", "install", path).CombinedOutput()
	if err != nil {
		Warn("Failed to run 'go install' for %s: %s", path, string(out))
	}
	return err
}

// buildAll builds all subpackages in the given path.
func buildAll(c cookoo.Context, path string) error {
	Info("Building all subpackages in %s\n", path)
	return buildPath(c, fmt.Sprintf("%s/...", path))
	/*
	if err := os.Chdir(path); err != nil {
		Warn("%s is not a directory. Skipping.\n", path)
	}

	out, err := exec.Command("go", "build", "./...").CombinedOutput()
	if err != nil {
		Warn("Failed to run 'go build' for %s: %s", path, string(out))
	}
	return err
	*/
}
