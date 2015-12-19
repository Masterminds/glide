package action

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
)

func Rebuild(yamlpath string) {
	conf := EnsureConfig(yamlpath)
	vpath, err := VendorPath(c)
	if err != nil {
		msg.Die("Could not get vendor path: %s", err)
	}

	msg.Info("Building dependencies.\n")

	if len(conf.Imports) == 0 {
		msg.Info("No dependencies found. Nothing built.\n")
		return
	}

	for _, dep := range conf.Imports {
		if err := buildDep(c, dep, vpath); err != nil {
			msg.Warn("Failed to build %s: %s\n", dep.Name, err)
		}
	}

	return true, nil
}

func buildDep(c cookoo.Context, dep *cfg.Dependency, vpath string) error {
	if len(dep.Subpackages) == 0 {
		buildPath(c, dep.Name)
	}

	for _, pkg := range dep.Subpackages {
		if pkg == "**" || pkg == "..." {
			//Info("Building all packages in %s\n", dep.Name)
			buildPath(c, path.Join(dep.Name, "..."))
		} else {
			paths, err := resolvePackages(vpath, dep.Name, pkg)
			if err != nil {
				Warn("Error resolving packages: %s", err)
			}
			buildPaths(c, paths)
		}
	}

	return nil
}

func resolvePackages(vpath, pkg, subpkg string) ([]string, error) {
	sdir, _ := os.Getwd()
	if err := os.Chdir(filepath.Join(vpath, pkg, subpkg)); err != nil {
		return []string{}, err
	}
	defer os.Chdir(sdir)
	p, err := filepath.Glob(path.Join(vpath, pkg, subpkg))
	if err != nil {
		return []string{}, err
	}
	for k, v := range p {
		nv := strings.TrimPrefix(v, vpath)
		p[k] = strings.TrimPrefix(nv, string(filepath.Separator))
	}
	return p, nil
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
	// . in a filepath.Join is removed so it needs to be prepended separately.
	p := "." + string(filepath.Separator) + filepath.Join("vendor", path)
	out, err := exec.Command("go", "install", p).CombinedOutput()
	if err != nil {
		Warn("Failed to run 'go install' for %s: %s", path, string(out))
	}
	return err
}
