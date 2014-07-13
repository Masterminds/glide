 package cmd

 import (
	 "os/exec"
	 "strings"
	 "fmt"
 )

// GoGetVCS implements a VCS for 'go get'.
type GoGetVCS struct {}

func (g *GoGetVCS) Get(dep *Dependency) error {
	out, err := exec.Command("go", "get", dep.Name).CombinedOutput()
	if err != nil {
		fmt.Print(string(out))
		if strings.Contains(string(out), "no buildable Go source") {
			return nil
		}
	}
	return err
}

func (g *GoGetVCS) Update(dep *Dependency) error {
	out, err := exec.Command("go", "get", "-u", dep.Name).CombinedOutput()
	if err != nil {
		fmt.Print(string(out))
		if strings.Contains(string(out), "no buildable Go source") {
			return nil
		}
	}
	return err
}

func (g *GoGetVCS) Version(dep *Dependency) error {
	return fmt.Errorf("%s does not have a repository/VCS set. No way to set version.", dep.Name)
}
