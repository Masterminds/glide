package cmd

import (
	"fmt"
	"os/exec"
	"strings"
)

// GoGetVCS implements a VCS for 'go get'.
type GoGetVCS struct{}

// Get uses `go get` to retrieve a package. This is used for first time installs.
func (g *GoGetVCS) Get(dep *Dependency) error {
	out, err := exec.Command("go", "get", "-d", dep.Name).CombinedOutput()
	if err != nil {
		//fmt.Print(string(out))
		if strings.Contains(string(out), "no buildable Go source") {
			Info("Go Get: %s", out)
			return nil
		}
		Warn("Go Get: %s", out)
	}
	return err
}

// Update uses `go get` to update a package.
func (g *GoGetVCS) Update(dep *Dependency) error {
	out, err := exec.Command("go", "get", "-d", "-u", dep.Name).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "no buildable Go source") {
			Info("Go Get: %s", out)
			return nil
		}
		Warn("Go Get: %s", out)
	}
	return err
}

// Version implements the interface to get version information for a package. Being
// implemented by `go get` there isn't version information available so an error
// is returned.
func (g *GoGetVCS) Version(dep *Dependency) error {
	return fmt.Errorf("%s does not have a repository/VCS set. No way to set version", dep.Name)
}

// LastCommit always retuns "" for GoGet, which is not revision-aware.
func (g *GoGetVCS) LastCommit(dep *Dependency) (string, error) {
	return "", nil
}
