package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// BzrVCS describes the BZR version control backend.
type BzrVCS struct{}

// We're not big Bazaar users, so we don't know whether we got this right.
// If you can help, please submit patches.

func (b *BzrVCS) Get(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)
	//fmt.Printf("[INFO] Cloning %s into %s\n", dep.Repository, dest)
	Info("Bzr: ")
	out, err := exec.Command("bzr", "branch", dep.Repository, dest).CombinedOutput()
	fmt.Print(string(out))
	return err
}

func (b *BzrVCS) Update(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	if _, err := os.Stat(dest); err != nil {
		// Assume a new package?
		Info("Looks like %s is a new package. Cloning.", dep.Name)
		return exec.Command("bzr", "branch", dep.Repository, dest).Run()
	}

	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(dest)
	defer os.Chdir(oldDir)

	// Because we can't predict which branch we want to be on, and since
	// we want to set checkouts explicitly, we should probably fetch.
	out, err := exec.Command("bzr", "pull", "--overwrite").CombinedOutput()
	fmt.Print(string(out))
	return err
}

func (b *BzrVCS) Version(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(dest)
	defer os.Chdir(oldDir)

	if len(dep.Reference) > 0 {
		// Now get to the right reference.
		tag := fmt.Sprintf("tag:%s", dep.Reference)
		//if out, err := exec.Command("bzr", "checkout", "-r", tag, dep.Repository).CombinedOutput(); err != nil {
		if out, err := exec.Command("bzr", "revert", "-r", tag).CombinedOutput(); err != nil {
			fmt.Println(string(out))
			return err
		}
	}
	return nil
}

func (b *BzrVCS) LastCommit(dep *Dependency) (string, error) {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	oldDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	os.Chdir(dest)
	defer os.Chdir(oldDir)
	out, err := exec.Command("bzr", "revno").CombinedOutput()
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(string(out), " ", 2)

	revno := parts[0]
	return revno, nil
}
