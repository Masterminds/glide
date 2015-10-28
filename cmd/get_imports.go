package cmd

import (
	"fmt"
	"sort"
	//"log"

	"os"
	"path"
	"runtime"
	"strings"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/util"
	"github.com/Masterminds/glide/yaml"
	"github.com/Masterminds/semver"
	v "github.com/Masterminds/vcs"
)

//func init() {
// Uncomment the line below and the log import to see the output
// from the vcs commands executed for each project.
//v.Logger = log.New(os.Stdout, "go-vcs", log.LstdFlags)
//}

// GetAll gets zero or more repos.
//
// This takes a package name, normalizes it, finds the repo, and installs it.
// It's the workhorse behind `glide get`.
//
// Params:
//	- packages ([]string): Package names to get.
// 	- verbose (bool): default false
//
// Returns:
// 	- []*Dependency: A list of constructed dependencies.
func GetAll(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	names := p.Get("packages", []string{}).([]string)
	cfg := p.Get("conf", nil).(*yaml.Config)
	insecure := p.Get("insecure", false).(bool)

	Info("Preparing to install %d package.", len(names))

	deps := []*yaml.Dependency{}
	for _, name := range names {
		cwd, err := VendorPath(c)
		if err != nil {
			return nil, err
		}

		root := util.GetRootFromPackage(name)
		if len(root) == 0 {
			return nil, fmt.Errorf("Package name is required for %q.", name)
		}

		if cfg.HasDependency(root) {
			Warn("Package %q is already in glide.yaml. Skipping", root)
			continue
		}

		dest := path.Join(cwd, root)

		var repoURL string
		if insecure {
			repoURL = "http://" + root
		} else {
			repoURL = "https://" + root
		}
		repo, err := v.NewRepo(repoURL, dest)
		if err != nil {
			Error("Could not construct repo for %q: %s", name, err)
			return false, err
		}

		dep := &yaml.Dependency{
			Name: root,
		}

		// When retriving from an insecure location set the repo to the
		// insecure location.
		if insecure {
			dep.Repository = "http://" + root
		}

		subpkg := strings.TrimPrefix(name, root)
		if len(subpkg) > 0 && subpkg != "/" {
			dep.Subpackages = []string{subpkg}
		}

		if err := repo.Get(); err != nil {
			return dep, err
		}

		cfg.Imports = append(cfg.Imports, dep)

		deps = append(deps, dep)

	}
	return deps, nil
}

// GetImports iterates over the imported packages and gets them.
func GetImports(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*yaml.Config)
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
			Warn("Skipped getting %s: %v\n", dep.Name, err)
		}
	}

	return true, nil
}

// UpdateImports iterates over the imported packages and updates them.
//
// Params:
//
// 	- force (bool): force packages to update (default false)
//	- conf (*yaml.Config): The configuration
// 	- packages([]string): The packages to update. Default is all.
func UpdateImports(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*yaml.Config)
	force := p.Get("force", true).(bool)
	plist := p.Get("packages", []string{}).([]string)
	pkgs := list2map(plist)
	restrict := len(pkgs) > 0

	cwd, err := VendorPath(c)
	if err != nil {
		return false, err
	}

	if len(cfg.Imports) == 0 {
		Info("No dependencies found. Nothing updated.\n")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if restrict && !pkgs[dep.Name] {
			Debug("===> Skipping %q", dep.Name)
			continue
		}

		// Hack: The updateCache global keeps us from re-updating the same
		// dependencies when we're recursing. We cache here to prevent
		// flattening from causing unnecessary updates.
		updateCache[dep.Name] = true

		if err := VcsUpdate(dep, cwd, force); err != nil {
			Warn("Update failed for %s: %s\n", dep.Name, err)
		}
	}

	return true, nil
}

// SetReference is a command to set the VCS reference (commit id, tag, etc) for
// a project.
func SetReference(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cfg := p.Get("conf", nil).(*yaml.Config)
	cwd, err := VendorPath(c)
	if err != nil {
		return false, err
	}

	if len(cfg.Imports) == 0 {
		Info("No references set.\n")
		return false, nil
	}

	for _, dep := range cfg.Imports {
		if err := VcsVersion(dep, cwd); err != nil {
			Warn("Failed to set version on %s to %s: %s\n", dep.Name, dep.Reference, err)
		}
	}

	return true, nil
}

// filterArchOs indicates a dependency should be filtered out because it is
// the wrong GOOS or GOARCH.
func filterArchOs(dep *yaml.Dependency) bool {
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

// VcsExists checks if the directory has a local VCS checkout.
func VcsExists(dep *yaml.Dependency, dest string) bool {
	repo, err := dep.GetRepo(dest)
	if err != nil {
		return false
	}

	return repo.CheckLocal()
}

// VcsGet figures out how to fetch a dependency, and then gets it.
//
// VcsGet installs into the dest.
func VcsGet(dep *yaml.Dependency, dest string) error {

	repo, err := dep.GetRepo(dest)
	if err != nil {
		return err
	}

	return repo.Get()
}

// VcsUpdate updates to a particular checkout based on the VCS setting.
func VcsUpdate(dep *yaml.Dependency, vend string, force bool) error {
	Info("Fetching updates for %s.\n", dep.Name)

	if filterArchOs(dep) {
		Info("%s is not used for %s/%s.\n", dep.Name, runtime.GOOS, runtime.GOARCH)
		return nil
	}

	dest := path.Join(vend, dep.Name)
	// If destination doesn't exist we need to perform an initial checkout.
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		if err = VcsGet(dep, dest); err != nil {
			Warn("Unable to checkout %s\n", dep.Name)
			return err
		}
	} else {
		// At this point we have a directory for the package.

		// When the directory is not empty and has no VCS directory it's
		// a vendored files situation.
		empty, err := isDirectoryEmpty(dest)
		if err != nil {
			return err
		}
		_, err = v.DetectVcsFromFS(dest)
		if empty == false && err == v.ErrCannotDetectVCS {
			Warn("%s appears to be a vendored package. Unable to update. Consider the '--update-vendored' flag.\n", dep.Name)
		} else {
			repo, err := dep.GetRepo(dest)

			// Tried to checkout a repo to a path that does not work. Either the
			// type or endpoint has changed. Force is being passed in so the old
			// location can be removed and replaced with the new one.
			// Warning, any changes in the old location will be deleted.
			// TODO: Put dirty checking in on the existing local checkout.
			if (err == v.ErrWrongVCS || err == v.ErrWrongRemote) && force == true {
				var newRemote string
				if len(dep.Repository) > 0 {
					newRemote = dep.Repository
				} else {
					newRemote = "https://" + dep.Name
				}

				Warn("Replacing %s with contents from %s\n", dep.Name, newRemote)
				rerr := os.RemoveAll(dest)
				if rerr != nil {
					return rerr
				}
				if err = VcsGet(dep, dest); err != nil {
					Warn("Unable to checkout %s\n", dep.Name)
					return err
				}
			} else if err != nil {
				return err
			} else {
				// Check if the current version is a tag or commit id. If it is
				// and that version is already checked out we can skip updating
				// which is faster than going out to the Internet to perform
				// an update.
				if dep.Reference != "" {
					version, err := repo.Version()
					if err != nil {
						return err
					}
					ib, err := isBranch(dep.Reference, repo)
					if err != nil {
						return err
					}

					// If the current version equals the ref and it's not a
					// branch it's a tag or commit id so we can skip
					// performing an update.
					if version == dep.Reference && !ib {
						Info("%s is already set to version %s. Skipping update.", dep.Name, dep.Reference)
						return nil
					}
				}

				if err := repo.Update(); err != nil {
					Warn("Download failed.\n")
					return err
				}
			}
		}
	}

	return nil
}

// VcsVersion set the VCS version for a checkout.
func VcsVersion(dep *yaml.Dependency, vend string) error {
	// If there is no refernece configured there is nothing to set.
	if dep.Reference == "" {
		return nil
	}

	cwd := path.Join(vend, dep.Name)

	// When the directory is not empty and has no VCS directory it's
	// a vendored files situation.
	empty, err := isDirectoryEmpty(cwd)
	if err != nil {
		return err
	}
	_, err = v.DetectVcsFromFS(cwd)
	if empty == false && err == v.ErrCannotDetectVCS {
		Warn("%s appears to be a vendored package. Unable to set new version. Consider the '--update-vendored' flag.\n", dep.Name)
	} else {
		repo, err := dep.GetRepo(cwd)
		if err != nil {
			return err
		}

		ver := dep.Reference
		// Referenes in Git can begin with a ^ which is similar to semver.
		// If there is a ^ prefix we assume it's a semver constraint rather than
		// part of the git/VCS commit id.
		if repo.IsReference(ver) && !strings.HasPrefix(ver, "^") {
			Info("Setting version for %s to %s.\n", dep.Name, ver)
		} else {

			// Create the constraing first to make sure it's valid before
			// working on the repo.
			constraint, err := semver.NewConstraint(ver)

			// Make sure the constriant is valid. At this point it's not a valid
			// reference so if it's not a valid constrint we can exit early.
			if err != nil {
				Warn("The reference '%s' is not valid\n", ver)
				return err
			}

			// Get the tags and branches (in that order)
			refs, err := getAllVcsRefs(repo)
			if err != nil {
				return err
			}

			// Convert and filter the list to semver.Version instances
			semvers := getSemVers(refs)

			// Sort semver list
			sort.Sort(sort.Reverse(semver.Collection(semvers)))
			found := false
			for _, v := range semvers {
				if constraint.Check(v) {
					found = true
					// If the constrint passes get the original reference
					ver = v.Original()
					break
				}
			}
			if found {
				Info("Detected semantic version. Setting version for %s to %s.\n", dep.Name, ver)
			} else {
				Warn("Unable to find semantic version for constraint %s %s\n", dep.Name, ver)
			}
		}
		if err := repo.UpdateVersion(ver); err != nil {
			Error("Failed to set version to %s: %s\n", dep.Reference, err)
			return err
		}
	}

	return nil
}

// VcsLastCommit gets the last commit ID from the given dependency.
func VcsLastCommit(dep *yaml.Dependency, vend string) (string, error) {
	cwd := path.Join(vend, dep.Name)
	repo, err := dep.GetRepo(cwd)
	if err != nil {
		return "", err
	}

	if repo.CheckLocal() == false {
		return "", fmt.Errorf("%s is not a VCS repo\n", dep.Name)
	}

	version, err := repo.Version()
	if err != nil {
		return "", err
	}

	return version, nil
}
