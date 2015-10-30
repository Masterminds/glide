package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
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
	home := p.Get("home", "").(string)
	cache := p.Get("cache", false).(bool)
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
		if err := VcsGet(dep, dest, home, cache); err != nil {
			return dep, err
		}

		cfg.Imports = append(cfg.Imports, dep)

		deps = append(deps, dep)

	}
	return deps, nil
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
	home := p.Get("home", "").(string)
	cache := p.Get("cache", false).(bool)
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

		if err := VcsUpdate(dep, cwd, home, force, cache); err != nil {
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
func VcsGet(dep *yaml.Dependency, dest, home string, cache bool) error {

	if !cache {
		// Check if the $GOPATH has a viable version to use and if so copy to vendor
		gps := Gopaths()
		for _, p := range gps {
			d := filepath.Join(p, "src", dep.Name)
			if _, err := os.Stat(d); err == nil {
				empty, err := isDirectoryEmpty(d)
				if empty || err != nil {
					continue
				}

				repo, err := dep.GetRepo(d)
				if err != nil {
					continue
				}

				// Dirty repos have uncomitted changes.
				if repo.IsDirty() {
					continue
				}

				// Having found a repo we copy it to vendor and update it.
				Debug("Found %s in GOPATH at %s. Copying to %s", dep.Name, d, dest)
				err = copyDir(d, dest)
				if err != nil {
					return err
				}

				// Update the repo in the vendor directory
				Debug("Updating %s, now in the vendor path at %s", dep.Name, dest)
				repo, err = dep.GetRepo(dest)
				if err != nil {
					return err
				}
				err = repo.Update()
				if err != nil {
					return err
				}

				// If there is no reference set on the dep we try to checkout
				// the default branch.
				if dep.Reference == "" {
					db := defaultBranch(repo, home)
					if db != "" {
						err = repo.UpdateVersion(db)
						if err != nil {
							Debug("Attempting to set the version on %s to %s failed. Error %s", dep.Name, db, err)
						}
					}
				}
				return nil
			}
		}

		// Since we didn't find an existing copy in the GOPATHs try to clone there.
		gp := Gopath()
		if gp != "" {
			d := filepath.Join(gp, "src", dep.Name)
			if _, err := os.Stat(d); os.IsNotExist(err) {
				// Empty directory so we checkout out the code here.
				Debug("Retrieving %s to %s before copying to vendor", dep.Name, d)
				repo, err := dep.GetRepo(d)
				if err != nil {
					return err
				}
				repo.Get()

				branch := findCurrentBranch(repo)
				if branch != "" {
					// we know the default branch so we can store it in the cache
					var loc string
					if dep.Repository != "" {
						loc = dep.Repository
					} else {
						loc = "https://" + dep.Name
					}
					key, err := cacheCreateKey(loc)
					if err == nil {
						Debug("Saving default branch for %s", repo.Remote())
						c := cacheRepoInfo{DefaultBranch: branch}
						saveCacheRepoData(key, c, home)
					}
				}

				Debug("Copying %s from GOPATH at %s to %s", dep.Name, d, dest)
				err = copyDir(d, dest)
				if err != nil {
					return err
				}

				return nil
			}
		}
	}

	// Check if the cache has a viable version and try to use that.
	var loc string
	if dep.Repository != "" {
		loc = dep.Repository
	} else {
		loc = "https://" + dep.Name
	}
	key, err := cacheCreateKey(loc)
	if err == nil {
		d := filepath.Join(home, "cache", "src", key)

		repo, err := dep.GetRepo(d)
		if err != nil {
			return err
		}
		// If the directory does not exist this is a first cache.
		if _, err = os.Stat(d); os.IsNotExist(err) {
			Debug("Adding %s to the cache for the first time", dep.Name)
			err = repo.Get()
			if err != nil {
				return err
			}
			branch := findCurrentBranch(repo)
			if branch != "" {
				// we know the default branch so we can store it in the cache
				var loc string
				if dep.Repository != "" {
					loc = dep.Repository
				} else {
					loc = "https://" + dep.Name
				}
				key, err := cacheCreateKey(loc)
				if err == nil {
					Debug("Saving default branch for %s", repo.Remote())
					c := cacheRepoInfo{DefaultBranch: branch}
					err = saveCacheRepoData(key, c, home)
					if err != nil {
						Debug("Error saving %s to cache. Error: %s", repo.Remote(), err)
					}
				}
			}

		} else {
			Debug("Updating %s in the cache", dep.Name)
			err = repo.Update()
			if err != nil {
				return err
			}
		}

		Debug("Copying %s from the cache to %s", dep.Name, dest)
		err = copyDir(d, dest)
		if err != nil {
			return err
		}

		return nil
	} else {
		Warn("Cache key generation error: %s", err)
	}

	// If unable to cache pull directly into the vendor/ directory.
	repo, err := dep.GetRepo(dest)
	if err != nil {
		return err
	}

	return repo.Get()
}

// VcsUpdate updates to a particular checkout based on the VCS setting.
func VcsUpdate(dep *yaml.Dependency, vend, home string, force, cache bool) error {
	Info("Fetching updates for %s.\n", dep.Name)

	if filterArchOs(dep) {
		Info("%s is not used for %s/%s.\n", dep.Name, runtime.GOOS, runtime.GOARCH)
		return nil
	}

	dest := path.Join(vend, dep.Name)
	// If destination doesn't exist we need to perform an initial checkout.
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		if err = VcsGet(dep, dest, home, cache); err != nil {
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
				if err = VcsGet(dep, dest, home, cache); err != nil {
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

// Some repos will have multiple branches in them (e.g. Git) while others
// (e.g. Svn) will not.
// TODO(mattfarina): Add API calls to github, bitbucket, etc.
func defaultBranch(repo v.Repo, home string) string {

	// Svn and Bzr use different locations (paths or entire locations)
	// for branches so we won't have a default branch.
	if repo.Vcs() == v.Svn || repo.Vcs() == v.Bzr {
		return ""
	}

	// Check the cache for a value.
	key, kerr := cacheCreateKey(repo.Remote())
	var d cacheRepoInfo
	if kerr == nil {
		d, err := cacheRepoData(key, home)
		if err == nil {
			if d.DefaultBranch != "" {
				return d.DefaultBranch
			}
		}
	}

	// If we don't have it in the store try some APIs
	r := repo.Remote()
	u, err := url.Parse(r)
	if err != nil {
		return ""
	}
	if u.Scheme == "" {
		// Where there is no scheme we try urls like git@github.com:foo/bar
		r = strings.Replace(r, ":", "/", -1)
		r = "ssh://" + r
		u, err = url.Parse(r)
		if err != nil {
			return ""
		}
		u.Scheme = ""
	}
	if u.Host == "github.com" {
		parts := strings.Split(u.Path, "/")
		if len(parts) != 2 {
			return ""
		}
		api := fmt.Sprintf("https://api.github.com/repos/%s/%s", parts[0], parts[1])
		resp, err := http.Get(api)
		if err != nil {
			return ""
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 || resp.StatusCode < 200 {
			return ""
		}
		body, err := ioutil.ReadAll(resp.Body)
		var data interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			return ""
		}
		gh := data.(map[string]interface{})
		db := gh["default_branch"].(string)
		if kerr == nil {
			d.DefaultBranch = db
			saveCacheRepoData(key, d, home)
		}
		return db
	}

	if u.Host == "bitbucket.org" {
		parts := strings.Split(u.Path, "/")
		if len(parts) != 2 {
			return ""
		}
		api := fmt.Sprintf("https://bitbucket.org/api/1.0/repositories/%s/%s/main-branch/", parts[0], parts[1])
		resp, err := http.Get(api)
		if err != nil {
			return ""
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 || resp.StatusCode < 200 {
			return ""
		}
		body, err := ioutil.ReadAll(resp.Body)
		var data interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			return ""
		}
		bb := data.(map[string]interface{})
		db := bb["name"].(string)
		if kerr == nil {
			d.DefaultBranch = db
			saveCacheRepoData(key, d, home)
		}
		return db
	}

	return ""
}

// From a local repo find out the current branch name if there is one.
func findCurrentBranch(repo v.Repo) string {
	Debug("Attempting to find current branch for %s", repo.Remote())
	// Svn and Bzr don't have default branches.
	if repo.Vcs() == v.Svn || repo.Vcs() == v.Bzr {
		return ""
	}

	if repo.Vcs() == v.Git {
		c := exec.Command("git", "symbolic-ref", "--short", "HEAD")
		c.Dir = repo.LocalPath()
		c.Env = envForDir(c.Dir)
		out, err := c.CombinedOutput()
		if err != nil {
			Debug("Unable to find current branch for %s, error: %s", repo.Remote(), err)
			return ""
		}
		return strings.TrimSpace(string(out))
	}

	if repo.Vcs() == v.Hg {
		c := exec.Command("hg", "branch")
		c.Dir = repo.LocalPath()
		c.Env = envForDir(c.Dir)
		out, err := c.CombinedOutput()
		if err != nil {
			Debug("Unable to find current branch for %s, error: %s", repo.Remote(), err)
			return ""
		}
		return strings.TrimSpace(string(out))
	}

	return ""
}

func envForDir(dir string) []string {
	env := os.Environ()
	return mergeEnvLists([]string{"PWD=" + dir}, env)
}

func mergeEnvLists(in, out []string) []string {
NextVar:
	for _, inkv := range in {
		k := strings.SplitAfterN(inkv, "=", 2)[0]
		for i, outkv := range out {
			if strings.HasPrefix(outkv, k) {
				out[i] = inkv
				continue NextVar
			}
		}
		out = append(out, inkv)
	}
	return out
}
