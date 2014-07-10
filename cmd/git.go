package cmd

import (
	"os"
	"os/exec"
	"fmt"
)

type GitVCS struct {}

// GitGet implements the getting logic for Git.
func (g *GitVCS) Get(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)
	fmt.Printf("[INFO] Cloning %s into %s\n", dep.Repository, dest)
	return exec.Command("git", "clone", dep.Repository, dest).Run()
}

func (g *GitVCS) Update(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	if _, err := os.Stat(dest); err != nil {
		// Assume a new package?
		fmt.Printf("[INFO] Looks like %s is a new package. Cloning.", dep.Name)
		return exec.Command("git", "clone", dep.Repository, dest).Run()
	}

	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(dest)
	defer os.Chdir(oldDir)

	// Because we can't predict which branch we want to be on, and since
	// we want to set checkouts explicitly, we should probably fetch.
	//out, err :=  exec.Command("git", "pull", "--ff-only").CombinedOutput()
	out, err :=  exec.Command("git", "fetch", "--all").CombinedOutput()
	fmt.Print(string(out))
	return err
}

func (g *GitVCS) Version(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(dest)
	defer os.Chdir(oldDir)

	//fmt.Printf("[INFO] Setting %s with 'git checkout'\n", dep.Name)

	// Now get to the right reference.
	if out, err := exec.Command("git", "checkout", dep.Reference).CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return err
	} else {
		updatedTo := "the latest"
		if dep.Reference != "" {
			updatedTo = dep.Reference
		}
		fmt.Printf("[INFO] Set version to %s to %s\n", dep.Name, updatedTo)
		//fmt.Print(string(out))
	}

	branchref := fmt.Sprintf("origin/%s", dep.Reference)
	err = exec.Command("git", "showref", "-q", branchref).Run()
	if err == nil {
		fmt.Printf("[DEBUG] Reference %s is to a branch.", dep.Reference)
		// git merge --ff-only origin $VERSION
		out, err := exec.Command("git", "merge", "--ff-only", "origin", dep.Reference).CombinedOutput()
		fmt.Println(out)
		if err != nil {
			return err
		}
	}

	return nil
}
