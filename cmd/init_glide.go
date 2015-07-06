package cmd

import (
	"fmt"
	"os"

	"github.com/Masterminds/cookoo"
)

var yamlTpl = `# Glide YAML configuration file
# Set this to your fully qualified package name, e.g.
# github.com/Masterminds/foo. This should be the
# top level package.
package: %s

# Declare your project's dependencies.
import:
  # Use 'go get' to fetch a package:
  #- package: github.com/Masterminds/cookoo
  # Get and manage a package with Git:
  #- package: github.com/Masterminds/cookoo
  #  # The repository URL
  #  repo: git@github.com:Masterminds/cookoo.git
  #  # A tag, branch, or SHA
  #  ref: 1.1.0
  #  # the VCS type (compare to bzr, hg, svn). You should
  #  # set this if you know it.
  #  vcs: git
`

// InitGlide initializes a new Glide project.
//
// Among other things, it creates a default glide.yaml.
//
// Params:
// 	- filename (string): The name of the glide YAML file. Default is glide.yaml.
// 	- project (string): The name of the project. Default is 'main'.
func InitGlide(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fname := p.Get("filename", "glide.yaml").(string)
	pname := p.Get("project", "main").(string)
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		Warn("If your GOPATH is automatically set by your shell, 'glide in' may not correctly set it.\n")
	}

	if _, err := os.Stat(fname); err == nil {
		cwd, _ := os.Getwd()
		return false, fmt.Errorf("Cowardly refusing to overwrite %s in %s", fname, cwd)
	}
	f, err := os.Create(fname)
	if err != nil {
		return false, err
	}
	defer f.Close()

	//f.WriteString(yamlTpl)
	fmt.Fprintf(f, yamlTpl, pname)

	if newgopath, err := GlideGopath(fname); err == nil {
		Info("Your new GOPATH is %s. Run 'glide gopath' to see it again.\n", newgopath)
	}
	Info("Initialized. You can now edit '%s'\n", fname)
	return true, nil
}
