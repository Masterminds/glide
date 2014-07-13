package cmd

import (
	"github.com/Masterminds/cookoo"
	"fmt"
	"os"
)

const (
	NoVCS uint = iota
	Git
	Bzr
	Hg
	Svn
)

func GetImports(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {

	cfg := p.Get("conf", nil).(*Config)

	if len(cfg.Imports) == 0 {
		Info("No dependencies found. Nothing downloaded.")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if err := VcsGet(dep); err != nil {
			Warn("Skipped getting %s: %s\n", dep.Name, err)
		}
	}

	return true, nil
}

func UpdateImports(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)

	if len(cfg.Imports) == 0 {
		Info("No dependencies found. Nothing updated.")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if err := VcsUpdate(dep); err != nil {
			Warn("Update failed for %s: %s\n", dep.Name, err)
		}
	}

	return true, nil
}

func CowardMode(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	gopath := os.Getenv("GOPATH")
	if len(gopath) == 0 {
		return false, fmt.Errorf("No GOPATH is set.")
	}

	if _, err := os.Stat(gopath); err != nil {
		return false, fmt.Errorf("Did you forget to 'glide install'? GOPATH=%s seems not to exist: %s", gopath, err)
	}

	ggpath := os.Getenv("GLIDE_GOPATH")
	if len(ggpath) > 0 && ggpath != gopath {
		Warn("Your GOPATH is set to %s, and we expected %s\n", gopath, ggpath)
	}
	return true, nil
}

func SetReference(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)

	if len(cfg.Imports) == 0 {
		Info("No dependencies found.")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if err := VcsVersion(dep); err != nil {
			Warn("Failed to set version on %s to %s: %s\n", dep.Name, dep.Reference, err)
		}
	}

	return true, nil
}

type VCS interface {
	Get(*Dependency) error
	Update(*Dependency) error
	Version(*Dependency) error
}

var (
	goGet VCS = new(GoGetVCS)
	git VCS = new(GitVCS)
	svn VCS = new(SvnVCS)
	bzr VCS = new(BzrVCS)
	hg VCS = new(HgVCS)
)

// VcsGet figures out how to fetch a dependency, and then gets it.
//
// Usually it delegates to lower-level *Get functions.
//
// See https://code.google.com/p/go/source/browse/src/cmd/go/vcs.go
func VcsGet(dep *Dependency) error {
	if dep.Repository == "" {
		Info("Installing %s with 'go get'\n", dep.Name)
		return goGet.Get(dep)
	}

	switch dep.VcsType {
	case Git:
		Info("Installing %s with Git (From %s)\n", dep.Name, dep.Repository)
		return git.Get(dep)
	case Bzr:
		Info("Installing %s with Bzr (From %s)\n", dep.Name, dep.Repository)
		return bzr.Get(dep)
	case Hg:
		Info("Installing %s with Hg (From %s)\n", dep.Name, dep.Repository)
		return hg.Get(dep)
	case Svn:
		Info("Installing %s with Svn (From %s)\n", dep.Name, dep.Repository)
		return svn.Get(dep)
	default:
		Warn("No handler for %s. Falling back to 'go get'.\n", dep.VcsType)
		return goGet.Get(dep)
	}
}

func VcsUpdate(dep *Dependency) error {
	// If no repository is set, we assume that the user wants us to use
	// 'go get'.
	if dep.Repository == "" {
		Info("Updating %s with 'go get -u'\n", dep.Name)
		return goGet.Update(dep)
	}

	switch dep.VcsType {
	case Git:
		Info("Updating %s with Git (From %s)\n", dep.Name, dep.Repository)
		return git.Update(dep)
	case Bzr:
		Info("Updating %s with Bzr (From %s)\n", dep.Name, dep.Repository)
		return bzr.Update(dep)
	case Hg:
		Info("Updating %s with Hg (From %s)\n", dep.Name, dep.Repository)
		return hg.Update(dep)
	case Svn:
		Info("Updating %s with Svn (From %s)\n", dep.Name, dep.Repository)
		return svn.Update(dep)
	default:
		Warn("No handler for %s. Falling back to 'go get -u'.\n", dep.VcsType)
		return goGet.Update(dep)
	}
}

func VcsVersion(dep *Dependency) error {
	if dep.VcsType == NoVCS {
		dep.VcsType, _ = GuessVCS(dep)
	}

	switch dep.VcsType {
	case Git:
		return git.Version(dep)
	case Bzr:
		return bzr.Version(dep)
	case Hg:
		return hg.Version(dep)
	case Svn:
		return svn.Version(dep)
	default:
		if len(dep.Reference) > 0 {
			Warn("Cannot update %s to specific version with VCS %d.\n", dep.Name, dep.VcsType)
			return goGet.Version(dep)
		}
		return nil
	}

}

func VcsSetReference(dep *Dependency) error {
	Warn("Cannot set reference. not implemented.\n")
	return nil
}


func GuessVCS(dep *Dependency) (uint, error) {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)

	if _, err := os.Stat(dest + "/.git"); err == nil {
		Info("Looks like %s is a Git repo.\n", dest)
		return Git, nil
	} else if _, err := os.Stat(dest + "/.bzr"); err == nil {
		Info("Looks like %s is a Bzr repo.\n", dest)
		return Bzr, nil
	} else if _, err := os.Stat(dest + "/.hg"); err == nil {
		Info("Looks like %s is a Mercurial repo.\n", dest)
		return Hg, nil
	} else if _, err := os.Stat(dest + "/.svn"); err == nil {
		Info("Looks like %s is a Subversion repo.\n", dest)
		return Svn, nil
	} else {
		return NoVCS, nil
	}
}
