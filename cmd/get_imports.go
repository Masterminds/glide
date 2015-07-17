package cmd

import (
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/Masterminds/cookoo"
	"golang.org/x/tools/go/vcs"
)

const (
	NoVCS = ""
	Git   = "git"
	Bzr   = "bzr"
	Hg    = "hg"
	Svn   = "svn"
)

// Get fetches a single package and puts it in vendor/.
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

	dep := &Dependency{
		Name:       repo.Root,
		VcsType:    repo.VCS.Cmd,
		Repository: repo.Repo,
	}
	subpkg := strings.TrimPrefix(name, repo.Root)
	if len(subpkg) > 0 && subpkg != "/" {
		dep.Subpackages = []string{subpkg}
	}

	dest := path.Join(cwd, repo.Root)
	if err := repo.VCS.Create(dest, repo.Repo); err != nil {
		return dep, err
	}

	cfg.Imports = append(cfg.Imports, dep)

	return dep, nil
}

// GetImports iterates over the imported packages and gets them.
func GetImports(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)
	cwd, err := VendorPath(c)
	if err != nil {
		Error("Failed to prepare vendor directory: %s", err)
		return false, err
	}

	if len(cfg.Imports) == 0 {
		Info("No dependencies found. Nothing downloaded.\n")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if err := VcsGet(dep, cwd); err != nil {
			Warn("Skipped getting %s: %s\n", dep.Name, err)
		}
	}

	return true, nil
}

// UpdateImports iterates over the imported packages and updates them.
func UpdateImports(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)
	cwd, err := VendorPath(c)
	if err != nil {
		return false, err
	}

	if len(cfg.Imports) == 0 {
		Info("No dependencies found. Nothing updated.\n")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if err := VcsUpdate(dep, cwd); err != nil {
			Warn("Update failed for %s: %s\n", dep.Name, err)
		}
	}

	return true, nil
}

func SetReference(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*Config)
	cwd, err := VendorPath(c)
	if err != nil {
		return false, err
	}

	if len(cfg.Imports) == 0 {
		Info("No dependencies found.\n")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if err := VcsVersion(dep, cwd); err != nil {
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
// VcsGet installs into the path at toPath.
func VcsGet(dep *Dependency, toPath string) error {

	if filterArchOs(dep) {
		Info("Ignoring %s for OS/ARch %s/%s", dep.Name, runtime.GOOS, runtime.GOARCH)
		return nil
	}

	cmd, err := dep.VCSCmd()
	if err != nil {
		Error("Could not resolve repository %s\n", dep.Name)
	}

	dest := path.Join(toPath, dep.Name)
	if err := cmd.Create(dest, dep.Name); err != nil {
		return err
	}

	if len(dep.Reference) > 0 {
		err := cmd.TagSync(dest, dep.Reference)
		if err != nil {
			Error("Failed to set revision: %s", err)
			return err
		}
	}
	return nil
}

// VcsUpdate updates to a particular checkout based on the VCS setting.
func VcsUpdate(dep *Dependency, vend string) error {
	Info("Fetching updates for %s.\n", dep.Name)

	if filterArchOs(dep) {
		Info("%s is not used for %s/%s.\n", dep.Name, runtime.GOOS, runtime.GOARCH)
		return nil
	}

	cmd, err := dep.VCSCmd()
	if err != nil {
		return err
	}

	dest := path.Join(vend, dep.Name)
	if err := cmd.Download(dest); err != nil {
		Warn("Download failed.\n")
		return err
	}

	/*
		if len(dep.Reference) > 0 {
			if err := cmd.TagSync(dest, dep.Reference); err != nil {
				Warn("Failed to set reference to %s. But source was downloaded. You may try to manually fix.\n", dep.Reference, err)
				return err
			}
		}
	*/
	return nil
}

func VcsVersion(dep *Dependency, vend string) error {
	Info("Setting version for %s.\n", dep.Name)

	cwd := path.Join(vend, dep.Name)
	cmd, err := dep.VCSCmd()
	if err != nil {
		return err
	}

	if err := cmd.TagSync(cwd, dep.Reference); err != nil {
		Error("Failed to sync to %s: %s\n", dep.Reference, err)
		return err
	}

	if cmd.Cmd == "git" {
		Info("XXX: Implement history-since function.")
	}

	return nil
}

func VcsLastCommit(dep *Dependency) (string, error) {

	cmd, err := dep.VCSCmd()
	if err != nil {
		return "", err
	}

	switch cmd.Cmd {
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
			Error("Cannot update %s to specific version with VCS %d.\n", dep.Name, dep.VcsType)
		}
		return "", nil
	}
}

func VcsSetReference(dep *Dependency) error {
	Warn("Cannot set reference. not implemented.\n")
	return nil
}

// GuessVCS attempts to guess guess the VCS used by a package.
func GuessVCS(dep *Dependency) (string, error) {
	cmd, err := dep.VCSCmd()
	if err != nil {
		return "", err
	}
	return cmd.Cmd, nil
}
