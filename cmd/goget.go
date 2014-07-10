 package cmd

 import (
	 "os/exec"
	 "fmt"
 )

// GoGetVCS implements a VCS for 'go get'.
type GoGetVCS struct {}

func (g *GoGetVCS) Get(dep *Dependency) error {
	err := exec.Command("go", "get", dep.Name).Run()
	return err
}

func (g *GoGetVCS) Update(dep *Dependency) error {
	err := exec.Command("go", "get", "-u", dep.Name).Run()
	return err
}

func (g *GoGetVCS) Version(dep *Dependency) error {
	return fmt.Errorf("%s does not have a repository/VCS set. No way to set version.", dep.Name)
}
