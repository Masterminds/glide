package cmd

import (
	"github.com/Masterminds/cookoo"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	//"fmt"
)

// Rebuild runs 'go build' in a directory.
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

		if pkg == "**" || pkg == "..." {
			//Info("Building all packages in %s\n", dep.Name)
			buildPath(c, path.Join(dep.Name, "..."))
		} else {
			paths, err := resolvePackages(gopath, dep.Name, pkg)
			if err != nil {
				Warn("Error resolving packages: %s", err)
			}
			//buildPath(c, path.Join(dep.Name, pkg))
			buildPaths(c, paths)
		}
	}

	return nil
}

func resolvePackages(gopath, pkg, subpkg string) ([]string, error) {
	sdir, _ := os.Getwd()
	if err := os.Chdir(path.Join(gopath, "src")); err != nil {
		return []string{}, err
	}
	defer os.Chdir(sdir)

	return filepath.Glob(path.Join(pkg, subpkg))
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
	Info("Running go build %s\n", path)
	out, err := exec.Command("go", "install", path).CombinedOutput()
	if err != nil {
		Warn("Failed to run 'go install' for %s: %s", path, string(out))
	}
	return err
}
