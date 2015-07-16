package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/Masterminds/cookoo"
	"golang.org/x/tools/go/vcs"
)

const (
	NoVCS uint = iota
	Git
	Bzr
	Hg
	Svn
)

// Get fetches a single package using the default `go get`.
//
// Params:
//	- package (string): Name of the package to get.
// 	- verbose (bool): default false
//
// Returns:
// 	- *Dependency: A dependency describing this package.
func Get(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	name := p.Get("package", "").(string)
	cfg := p.Get("conf", nil).(*Config)
	verbose := p.Get("verbose", false).(bool)

	cwd, err := VendorPath(c)
	if err != nil {
		return nil, err
	}

	repo, err := vcs.RepoRootForImportPath(name, verbose)
	if err != nil {
		return nil, err
	}

	if cfg.HasDependency(repo.Root) {
		return nil, fmt.Errorf("Package '%s' is already in glide.yaml", repo.Root)
	}

	if len(repo.Root) == 0 {
		return nil, fmt.Errorf("Package name is required.")
	}

	dep := &Dependency{Name: repo.Root}
	subpkg := strings.TrimPrefix(name, repo.Root)
	if len(subpkg) > 0 && subpkg != "/" {
		dep.Subpackages = []string{subpkg}
	}

	dest := path.Join(cwd, repo.Root)
	if err := repo.VCS.Create(dest, repo.Repo); err != nil {
		return dep, err
	}

	/*
		if err := VcsGet(dep); err != nil {
			return dep, err
		}
	*/

	cfg.Imports = append(cfg.Imports, dep)

	return dep, nil
}

// GetImports iterates over the imported packages and gets them.
func GetImports(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)

	if len(cfg.Imports) == 0 {
		Info("No dependencies found. Nothing downloaded.\n")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if err := VcsGet(dep); err != nil {
			Warn("Skipped getting %s: %s\n", dep.Name, err)
		}
	}

	return true, nil
}

// UpdateImports iterates over the imported packages and updates them.
func UpdateImports(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)

	if len(cfg.Imports) == 0 {
		Info("No dependencies found. Nothing updated.\n")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if err := VcsUpdate(dep); err != nil {
			Warn("Update failed for %s: %s\n", dep.Name, err)
		}
	}

	return true, nil
}

// CowardMode checks that the environment is setup before continuing on. If not
// setup and error is returned.
func CowardMode(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	gopath := os.Getenv("GOPATH")
	if len(gopath) == 0 {
		return false, fmt.Errorf("No GOPATH is set.\n")
	}

	_, err := os.Stat(path.Join(gopath, "src"))
	if err != nil {
		Error("Could not find %s/src.\n", gopath)
		Info("As of Glide 0.5/Go 1.5, this is required.\n")
		return false, err
	}

	return true, nil
}

func SetReference(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)

	if len(cfg.Imports) == 0 {
		Info("No dependencies found.\n")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if err := VcsVersion(dep); err != nil {
			Warn("Failed to set version on %s to %s: %s\n", dep.Name, dep.Reference, err)
		}
	}

	return true, nil
}

var (
	goGet VCS = new(GoGetVCS)
	git   VCS = new(GitVCS)
	svn   VCS = new(SvnVCS)
	bzr   VCS = new(BzrVCS)
	hg    VCS = new(HgVCS)
)

// filterArchOs indicates a dependency should be filtered out because it is
// the wrong GOOS or GOARCH.
func filterArchOs(dep *Dependency) bool {
	found := false
	if len(dep.Arch) > 0 {
		for _, a := range dep.Arch {
			if a == runtime.GOARCH {
				found = true
			}
		}
		// If it's not found, it should be filtered out.
		if !found {
			return true
		}
	}

	found = false
	if len(dep.Os) > 0 {
		for _, o := range dep.Os {
			if o == runtime.GOOS {
				found = true
			}
		}
		if !found {
			return true
		}

	}

	return false
}

// VcsGet figures out how to fetch a dependency, and then gets it.
//
// Usually it delegates to lower-level *Get functions.
//
// See https://code.google.com/p/go/source/browse/src/cmd/go/vcs.go
func VcsGet(dep *Dependency) error {

	if filterArchOs(dep) {
		Info("Ignoring %s for OS/ARch %s/%s", dep.Name, runtime.GOOS, runtime.GOARCH)
		return nil
	}

	// See note in VcsUpdate.
	if dep.Repository == "" && dep.Reference == "" {
		Info("Installing %s with 'go get'\n", dep.Name)
		return goGet.Get(dep)
	}

	switch dep.VcsType {
	case Git:
		if dep.Repository == "" {
			dep.Repository = "https://" + dep.Name
		}
		Info("Installing %s with Git (From %s)\n", dep.Name, dep.Repository)
		return git.Get(dep)
	case Bzr:
		Info("Installing %s with Bzr (From %s)\n", dep.Name, dep.Repository)
		return bzr.Get(dep)
	case Hg:
		if dep.Repository == "" {
			dep.Repository = "https://" + dep.Name
		}
		Info("Installing %s with Hg (From %s)\n", dep.Name, dep.Repository)
		return hg.Get(dep)
	case Svn:
		Info("Installing %s with Svn (From %s)\n", dep.Name, dep.Repository)
		return svn.Get(dep)
	default:
		if dep.VcsType == NoVCS {
			Info("Defaulting to 'go get %s'\n", dep.Name)
			if len(dep.Reference) > 0 {
				Warn("Ref is set to %s, but no VCS is set. This can cause inconsistencies.\n", dep.Reference)
			}
		} else {
			Warn("No handler for %d. Falling back to 'go get %s'.\n", dep.VcsType, dep.Name)
		}
		return goGet.Get(dep)
	}
}

// VcsUpdate updates to a particular checkout based on the VCS setting.
func VcsUpdate(dep *Dependency) error {

	if filterArchOs(dep) {
		Info("%s is not used for %s/%s.\n", dep.Name, runtime.GOOS, runtime.GOARCH)
		return nil
	}

	// If there is no Ref set, if Repository is empty, and if the VCS not known
	// we should just default to Go Get.
	//
	// Why do we care if Ref is blank? As of Go 1.3, go get builds a .a
	// file for each library. But if we set a Ref, that will switch the source
	// code, but not necessarily build a .a file. So we want to make sure not
	// to default to 'go get' if we're then going to grab a specific version.
	if dep.Reference == "" && dep.Repository == "" && dep.VcsType == NoVCS {
		Info("Reference, repo, and VCS not set. Falling back to 'go get -u %s'.\n", dep.Name)
		return goGet.Update(dep)
	}

	if dep.Repository == "" && (dep.VcsType == Git || dep.VcsType == Hg) {
		dep.Repository = "https://" + dep.Name
	}

	if dep.VcsType == NoVCS {
		guess, err := GuessVCS(dep)
		if err != nil {
			Warn("Tried to guess VCS type, but failed: %s", err)
		} else {
			dep.VcsType = guess
		}
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
		if dep.VcsType == NoVCS {
			Info("No VCS set. Updating with 'go get -u %s'\n", dep.Name)
			if len(dep.Reference) > 0 {
				Warn("Ref is set to %s, but no VCS is set. This can cause inconsistencies.\n", dep.Reference)
			}
		} else {
			Warn("No handler for this repo type. Falling back to 'go get -u %s'.\n", dep.Name)
		}
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

func VcsLastCommit(dep *Dependency) (string, error) {
	if dep.VcsType == NoVCS {
		dep.VcsType, _ = GuessVCS(dep)
	}

	switch dep.VcsType {
	case Git:
		return git.LastCommit(dep)
	case Bzr:
		return bzr.LastCommit(dep)
	case Hg:
		return hg.LastCommit(dep)
	case Svn:
		return svn.LastCommit(dep)
	default:
		if len(dep.Reference) > 0 {
			Warn("Cannot update %s to specific version with VCS %d.\n", dep.Name, dep.VcsType)
			return goGet.LastCommit(dep)
		}
		return "", nil
	}
}

func VcsSetReference(dep *Dependency) error {
	Warn("Cannot set reference. not implemented.\n")
	return nil
}

// GuessVCS attempts to guess guess the VCS used by a package.
func GuessVCS(dep *Dependency) (uint, error) {
	dest := fmt.Sprintf("%s/src/%s", os.Getenv("GOPATH"), dep.Name)
	//Debug("Looking in %s for hints about VCS type.\n", dest)

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
