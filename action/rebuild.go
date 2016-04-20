package action

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

// Rebuild rebuilds '.a' files for a project.
//
// Prior to Go 1.4, this could substantially reduce time on incremental compiles.
// It remains to be seen whether this is tremendously beneficial to modern Go
// programs.
func Rebuild() {
	msg.Warn("The rebuild command is deprecated and will be removed in a future version")
	msg.Warn("Use the go install command instead")
	conf := EnsureConfig()
	vpath, err := gpath.Vendor()
	if err != nil {
		msg.Die("Could not get vendor path: %s", err)
	}

	msg.Info("Building dependencies.\n")

	if len(conf.Imports) == 0 {
		msg.Info("No dependencies found. Nothing built.\n")
		return
	}

	for _, dep := range conf.Imports {
		if err := buildDep(dep, vpath); err != nil {
			msg.Warn("Failed to build %s: %s\n", dep.Name, err)
		}
	}
}

func buildDep(dep *cfg.Dependency, vpath string) error {
	if len(dep.Subpackages) == 0 {
		buildPath(dep.Name)
	}

	for _, pkg := range dep.Subpackages {
		if pkg == "**" || pkg == "..." {
			//Info("Building all packages in %s\n", dep.Name)
			buildPath(path.Join(dep.Name, "..."))
		} else {
			paths, err := resolvePackages(vpath, dep.Name, pkg)
			if err != nil {
				msg.Warn("Error resolving packages: %s", err)
			}
			buildPaths(paths)
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

func buildPaths(paths []string) error {
	for _, path := range paths {
		if err := buildPath(path); err != nil {
			return err
		}
	}

	return nil
}

func buildPath(path string) error {
	msg.Info("Running go build %s\n", path)
	// . in a filepath.Join is removed so it needs to be prepended separately.
	p := "." + string(filepath.Separator) + filepath.Join("vendor", path)
	out, err := exec.Command(goExecutable(), "install", p).CombinedOutput()
	if err != nil {
		msg.Warn("Failed to run 'go install' for %s: %s", path, string(out))
	}
	return err
}

