package cmd

import (
	"os"
	"os/exec"
	"fmt"
	"strings"
)

type HgVCS struct {}

// If you can help clean this up or improve it, please submit patches!

// hgGet implements the getting logic for hg.
func (h *HgVCS) Get(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)
	//Info("Cloning %s into %s\n", dep.Repository, dest)
	fmt.Print("[INFO] hg: ")
	out, err := exec.Command("hg", "clone", "-U", dep.Repository, dest).CombinedOutput()
	fmt.Print(string(out))
	return err
}

func (h *HgVCS) Update(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	if _, err := os.Stat(dest); err != nil {
		// Assume a new package?
		Info("Looks like %s is a new package. Cloning.", dep.Name)
		return exec.Command("hg", "clone", "-U", dep.Repository, dest).Run()
	}

	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(dest)
	defer os.Chdir(oldDir)

	out, err :=  exec.Command("hg", "pull").CombinedOutput()
	fmt.Print(string(out))
	return err
}

func (h *HgVCS) Version(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(dest)
	defer os.Chdir(oldDir)

	//Info("Setting %s with 'hg checkout'\n", dep.Name)

	// Now get to the right reference.
	if len(dep.Reference) > 0 {
		if out, err := exec.Command("hg", "update", "-q", dep.Reference).CombinedOutput(); err != nil {
			fmt.Println(string(out))
			return err
		}
		Info("Set version to %s to %s\n", dep.Name, dep.Reference)
	}
	return nil
}

func (h *HgVCS) LastCommit(dep *Dependency) (string, error) {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	oldDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	os.Chdir(dest)
	defer os.Chdir(oldDir)

	out, err := exec.Command("hg", "identify").CombinedOutput()
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(string(out), " ", 2)

	sha := parts[0]
	return sha, nil
}
