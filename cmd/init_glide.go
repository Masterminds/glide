package cmd

import (
	"github.com/Masterminds/cookoo"
	"fmt"
	"os"
)

var yamlTpl = `# Glide YAML configuration file
# Set this to your fully qualified package name, e.g.
# github.com/Masterminds/foo. This should be the
# top level package.
package: main

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

func InitGlide(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {

	if gopath := os.Getenv("GOPATH"); gopath != "" {
		fmt.Printf("[WARN] If your GOPATH is automatically set by your shell, 'glide in' may not correctly set it.")
	}

	if _, err := os.Stat("./glide.yaml"); err == nil {
		cwd, _ := os.Getwd()
		return false, fmt.Errorf("Cowardly refusing to overwrite glide.yaml in %s", cwd)
	}
	f, err := os.Create("./glide.yaml")
	if err != nil {
		return false, err
	}
	defer f.Close()

	f.WriteString(yamlTpl)


	if newgopath, err := GlideGopath(); err == nil {
		fmt.Printf("[INFO] Your new GOPATH is %s. Run 'glide gopath' to see it again.\n", newgopath)
	}
	fmt.Printf("[INFO] Initialized. You can now edit 'glide.yaml'\n")
	return true, nil
}
