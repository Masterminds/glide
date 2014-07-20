package cmd

import (
	"os"
	"os/exec"
	"fmt"
	"strings"
	"regexp"
)

type GitVCS struct {}

// GitGet implements the getting logic for Git.
func (g *GitVCS) Get(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)
	//Info("Cloning %s into %s\n", dep.Repository, dest)
	Info("Git: ")
	out, err := exec.Command("git", "clone", dep.Repository, dest).CombinedOutput()
	fmt.Print(string(out))
	return err
}

func (g *GitVCS) Update(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	if _, err := os.Stat(dest); err != nil {
		// Assume a new package?
		Info("Looks like %s is a new package. Cloning.", dep.Name)
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
	Info("Git: ")
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

	//Info("Setting %s with 'git checkout'\n", dep.Name)

	// Now get to the right reference.
	if out, err := exec.Command("git", "checkout", dep.Reference).CombinedOutput(); err != nil {
		fmt.Println(string(out))
		return err
	} else {
		updatedTo := "the latest"
		if dep.Reference != "" {
			updatedTo = dep.Reference
		}
		Info("Set version to %s to %s\n", dep.Name, updatedTo)
		//fmt.Print(string(out))
	}

	branchref := fmt.Sprintf("origin/%s", dep.Reference)
	err = exec.Command("git", "showref", "-q", branchref).Run()
	if err == nil {
		Debug("Reference %s is to a branch.", dep.Reference)
		// git merge --ff-only origin $VERSION
		out, err := exec.Command("git", "merge", "--ff-only", "origin", dep.Reference).CombinedOutput()
		fmt.Println(out)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *GitVCS) LastCommit(dep *Dependency) (string, error) {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	oldDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	os.Chdir(dest)
	defer os.Chdir(oldDir)

	out, err := exec.Command("git", "log", "-n", "1", "--pretty=format:%h%d").CombinedOutput()
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(string(out), " ", 2)

	sha := parts[0]

	// Send back a tag if a tag matches.
	if len(parts) > 1 && strings.Contains(parts[1], "tag: ") {
		re := regexp.MustCompile("tag: ([^,)]*)")
		subs := re.FindStringSubmatch(parts[1])
		if len(subs) > 1 {
			return subs[1], nil
		}
	}

	return sha, nil
}

