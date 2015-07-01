package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SvnVCS implements the VCS interface for the Svn source control.
type SvnVCS struct{}

// If you can help clean this up or improve it, please submit patches!

// Get implements the getting logic for checking out a codebase from SVN.
func (s *SvnVCS) Get(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)
	//Info("Cloning %s into %s\n", dep.Repository, dest)
	fmt.Print("[INFO] svn: ")
	out, err := exec.Command("svn", "checkout", dep.Repository, dest).CombinedOutput()
	fmt.Print(string(out))
	return err
}

// Update performs an SVN update to an existing checkout.
func (s *SvnVCS) Update(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	if _, err := os.Stat(dest); err != nil {
		// Assume a new package?
		Info("Looks like %s is a new package. Cloning.", dep.Name)
		return exec.Command("svn", "checkout", dep.Repository, dest).Run()
	}

	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(dest)
	defer os.Chdir(oldDir)

	out, err := exec.Command("svn", "update").CombinedOutput()
	fmt.Print(string(out))
	return err
}

// Version sets the version of a package currently checked out via SVN. For
// more detail see the SetReference function.
func (s *SvnVCS) Version(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(dest)
	defer os.Chdir(oldDir)

	if len(dep.Reference) > 0 {
		if out, err := exec.Command("svn", "update", "-r", dep.Reference).CombinedOutput(); err != nil {
			fmt.Println(string(out))
			return err
		}
	}
	return nil
}

// LastCommit retrieves the current version.
func (s *SvnVCS) LastCommit(dep *Dependency) (string, error) {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	oldDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	os.Chdir(dest)
	defer os.Chdir(oldDir)
	out, err := exec.Command("svnversion", ".").CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
