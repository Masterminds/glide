package cmd

import (
	"github.com/Masterminds/cookoo"
	"fmt"
	"os"
	"os/exec"
)

func GetImports(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {

	cfg := p.Get("conf", nil).(*Config)

	if len(cfg.Imports) == 0 {
		fmt.Printf("[INFO] No dependencies found. Nothing downloaded.")
		return false, nil
	}

	for i, dep := range cfg.Imports {
		fmt.Printf("[INFO] %d: Getting %s\n", i, dep.Name)
		if err := VcsGet(dep); err != nil {
			fmt.Printf("[WARN] Skipped getting %s: %s\n", dep.Name, err)
		}
	}

	return nil, nil
}

// VcsGet figures out how to fetch a dependency, and then gets it.
//
// Usually it delegates to lower-level *Get functions.
func VcsGet(dep *Dependency) error {
	if dep.Repository == "" {
		return GoGet(dep)
	}

	switch dep.VcsType {
	case "git":
		return GitGet(dep)
	default:
		fmt.Printf("[WARN] No handler for %s. Falling back to 'go get'.\n", dep.VcsType)
		return GoGet(dep)
	}
}

func VcsSetReference(dep *Dependency) error {
	fmt.Printf("[WARN] Cannot set reference. not implemented.\n")
	return nil
}

func GoGet(dep *Dependency) error {
	err := exec.Command("go", "get", dep.Name).Run()
	return err
}

// GitGet implements the getting logic for Git.
func GitGet(dep *Dependency) error {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)
	fmt.Printf("[INFO] Cloning %s into %s\n", dep.Repository, dest)
	return exec.Command("git", "clone", dep.Repository, dest).Run()
}
