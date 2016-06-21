package repo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	cp "github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/semver"
	v "github.com/Masterminds/vcs"
)

// VcsUpdate updates to a particular checkout based on the VCS setting.
func VcsUpdate(dep *cfg.Dependency, dest, home string, cache, cacheGopath, useGopath, force, updateVendored bool, updated *UpdateTracker) error {

	// If the dependency has already been pinned we can skip it. This is a
	// faster path so we don't need to resolve it again.
	if dep.Pin != "" {
		msg.Debug("Dependency %s has already been pinned. Fetching updates skipped.", dep.Name)
		return nil
	}

	if updated.Check(dep.Name) {
		msg.Debug("%s was already updated, skipping.", dep.Name)
		return nil
	}
	updated.Add(dep.Name)

	msg.Info("--> Fetching updates for %s.", dep.Name)

	if filterArchOs(dep) {
		msg.Info("%s is not used for %s/%s.\n", dep.Name, runtime.GOOS, runtime.GOARCH)
		return nil
	}

	// If destination doesn't exist we need to perform an initial checkout.
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		if err = VcsGet(dep, dest, home, cache, cacheGopath, useGopath); err != nil {
			msg.Warn("Unable to checkout %s\n", dep.Name)
			return err
		}
	} else {
		// At this point we have a directory for the package.

		// When the directory is not empty and has no VCS directory it's
		// a vendored files situation.
		empty, err := gpath.IsDirectoryEmpty(dest)
		if err != nil {
			return err
		}
		_, err = v.DetectVcsFromFS(dest)
		if updateVendored == false && empty == false && err == v.ErrCannotDetectVCS {
			msg.Warn("%s appears to be a vendored package. Unable to update. Consider the '--update-vendored' flag.\n", dep.Name)
		} else if updateVendored == false && empty == true && err == v.ErrCannotDetectVCS {
			msg.Warn("%s is an empty directory. Fetching a new copy of the dependency.", dep.Name)
			msg.Debug("Removing empty directory %s", dest)
			err := os.RemoveAll(dest)
			if err != nil {
				return err
			}
			if err = VcsGet(dep, dest, home, cache, cacheGopath, useGopath); err != nil {
				msg.Warn("Unable to checkout %s\n", dep.Name)
				return err
			}
		} else {

			if updateVendored == true && empty == false && err == v.ErrCannotDetectVCS {
				// A vendored package, no repo, and updating the vendored packages
				// has been opted into.
				msg.Info("%s is a vendored package. Updating.", dep.Name)
				err = os.RemoveAll(dest)
				if err != nil {
					msg.Err("Unable to update vendored dependency %s.\n", dep.Name)
					return err
				}
				dep.UpdateAsVendored = true

				if err = VcsGet(dep, dest, home, cache, cacheGopath, useGopath); err != nil {
					msg.Warn("Unable to checkout %s\n", dep.Name)
					return err
				}

				return nil
			}

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

				msg.Warn("Replacing %s with contents from %s\n", dep.Name, newRemote)
				rerr := os.RemoveAll(dest)
				if rerr != nil {
					return rerr
				}
				if err = VcsGet(dep, dest, home, cache, cacheGopath, useGopath); err != nil {
					msg.Warn("Unable to checkout %s\n", dep.Name)
					return err
				}

				repo, err = dep.GetRepo(dest)
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else if repo.IsDirty() {
				return fmt.Errorf("%s contains uncommitted changes. Skipping update", dep.Name)
			}

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
					msg.Debug("%s is already set to version %s. Skipping update.", dep.Name, dep.Reference)
					return nil
				}
			}

			if err := repo.Update(); err != nil {
				msg.Warn("Download failed.\n")
				return err
			}
		}
	}

	return nil
}

// VcsVersion set the VCS version for a checkout.
func VcsVersion(dep *cfg.Dependency, vend string) error {

	// If the dependency has already been pinned we can skip it. This is a
	// faster path so we don't need to resolve it again.
	if dep.Pin != "" {
		msg.Debug("Dependency %s has already been pinned. Setting version skipped.", dep.Name)
		return nil
	}

	cwd := filepath.Join(vend, dep.Name)

	// If there is no reference configured there is nothing to set.
	if dep.Reference == "" {
		// Before exiting update the pinned version
		repo, err := dep.GetRepo(cwd)
		if err != nil {
			return err
		}
		dep.Pin, err = repo.Version()
		if err != nil {
			return err
		}
		return nil
	}

	// When the directory is not empty and has no VCS directory it's
	// a vendored files situation.
	empty, err := gpath.IsDirectoryEmpty(cwd)
	if err != nil {
		return err
	}
	_, err = v.DetectVcsFromFS(cwd)
	if empty == false && err == v.ErrCannotDetectVCS {
		msg.Warn("%s appears to be a vendored package. Unable to set new version. Consider the '--update-vendored' flag.\n", dep.Name)
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
			msg.Info("--> Setting version for %s to %s.\n", dep.Name, ver)
		} else {

			// Create the constraing first to make sure it's valid before
			// working on the repo.
			constraint, err := semver.NewConstraint(ver)

			// Make sure the constriant is valid. At this point it's not a valid
			// reference so if it's not a valid constrint we can exit early.
			if err != nil {
				msg.Warn("The reference '%s' is not valid\n", ver)
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
				msg.Info("--> Detected semantic version. Setting version for %s to %s.", dep.Name, ver)
			} else {
				msg.Warn("--> Unable to find semantic version for constraint %s %s", dep.Name, ver)
			}
		}
		if err := repo.UpdateVersion(ver); err != nil {
			return err
		}
		dep.Pin, err = repo.Version()
		if err != nil {
			return err
		}
	}

	return nil
}

// VcsGet figures out how to fetch a dependency, and then gets it.
//
// VcsGet installs into the dest.
func VcsGet(dep *cfg.Dependency, dest, home string, cache, cacheGopath, useGopath bool) error {
	// When not skipping the $GOPATH look in it for a copy of the package
	if useGopath {
		// Check if the $GOPATH has a viable version to use and if so copy to vendor
		gps := gpath.Gopaths()
		for _, p := range gps {
			d := filepath.Join(p, "src", dep.Name)
			if _, err := os.Stat(d); err == nil {
				empty, err := gpath.IsDirectoryEmpty(d)
				if empty || err != nil {
					continue
				}

				repo, err := dep.GetRepo(d)
				if err != nil {
					continue
				}

				// Dirty repos have uncommitted changes.
				if repo.IsDirty() {
					continue
				}

				// Having found a repo we copy it to vendor and update it.
				msg.Info("Copying package %s from the GOPATH.", dep.Name)
				msg.Debug("Found %s in GOPATH at %s. Copying to %s", dep.Name, d, dest)
				err = gpath.CopyDir(d, dest)
				if err != nil {
					return err
				}

				// Update the repo in the vendor directory
				msg.Debug("Updating %s, now in the vendor path at %s", dep.Name, dest)
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
						if err != nil && msg.Default.IsDebugging {
							msg.Debug("Attempting to set the version on %s to %s failed. Error %s", dep.Name, db, err)
						}
					}
				}
				return nil
			}
		}
	}

	// When opting in to cache in the GOPATH attempt to do put a copy there.
	if cacheGopath {

		// Since we didn't find an existing copy in the GOPATHs try to clone there.
		gp := gpath.Gopath()
		if gp != "" {
			d := filepath.Join(gp, "src", dep.Name)
			if _, err := os.Stat(d); os.IsNotExist(err) {
				// Empty directory so we checkout out the code here.
				msg.Debug("Retrieving %s to %s before copying to vendor", dep.Name, d)
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
					key, err := cp.Key(loc)
					if err == nil {
						msg.Debug("Saving default branch for %s", repo.Remote())
						c := cp.RepoInfo{DefaultBranch: branch}
						err = cp.SaveRepoData(key, c)
						if msg.Default.IsDebugging && err == cp.ErrCacheDisabled {
							msg.Debug("Unable to cache default branch because caching is disabled")
						}
					}
				}

				msg.Debug("Copying %s from GOPATH at %s to %s", dep.Name, d, dest)
				err = gpath.CopyDir(d, dest)
				if err != nil {
					return err
				}

				return nil
			}
		}
	}

	// If opting in to caching attempt to put it in the cache folder
	if cache {
		// Check if the cache has a viable version and try to use that.
		var loc string
		if dep.Repository != "" {
			loc = dep.Repository
		} else {
			loc = "https://" + dep.Name
		}
		key, err := cp.Key(loc)
		if err == nil {
			location, err := cp.Location()
			if err != nil {
				return err
			}
			d := filepath.Join(location, "src", key)

			repo, err := dep.GetRepo(d)
			if err != nil {
				return err
			}
			// If the directory does not exist this is a first cache.
			if _, err = os.Stat(d); os.IsNotExist(err) {
				msg.Debug("Adding %s to the cache for the first time", dep.Name)
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
					key, err := cp.Key(loc)
					if err == nil {
						msg.Debug("Saving default branch for %s", repo.Remote())
						c := cp.RepoInfo{DefaultBranch: branch}
						err = cp.SaveRepoData(key, c)
						if err == cp.ErrCacheDisabled {
							msg.Debug("Unable to cache default branch because caching is disabled")
						} else if err != nil {
							msg.Debug("Error saving %s to cache. Error: %s", repo.Remote(), err)
						}
					}
				}

			} else {
				msg.Debug("Updating %s in the cache", dep.Name)
				err = repo.Update()
				if err != nil {
					return err
				}
			}

			msg.Debug("Copying %s from the cache to %s", dep.Name, dest)
			err = gpath.CopyDir(d, dest)
			if err != nil {
				return err
			}

			return nil
		}

		msg.Warn("Cache key generation error: %s", err)
	}

	// If unable to cache pull directly into the vendor/ directory.
	repo, err := dep.GetRepo(dest)
	if err != nil {
		return err
	}

	gerr := repo.Get()

	// Attempt to cache the default branch
	if cache {
		if branch := findCurrentBranch(repo); branch != "" {
			// we know the default branch so we can store it in the cache
			var loc string
			if dep.Repository != "" {
				loc = dep.Repository
			} else {
				loc = "https://" + dep.Name
			}
			key, err := cp.Key(loc)
			if err == nil {
				msg.Debug("Saving default branch for %s", repo.Remote())
				c := cp.RepoInfo{DefaultBranch: branch}
				err = cp.SaveRepoData(key, c)
				if err == cp.ErrCacheDisabled {
					msg.Debug("Unable to cache default branch because caching is disabled")
				} else if err != nil {
					msg.Debug("Error saving %s to cache - Error: %s", repo.Remote(), err)
				}
			}
		}
	}

	return gerr
}

// filterArchOs indicates a dependency should be filtered out because it is
// the wrong GOOS or GOARCH.
//
// FIXME: Should this be moved to the dependency package?
func filterArchOs(dep *cfg.Dependency) bool {
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

// isBranch returns true if the given string is a branch in VCS.
func isBranch(branch string, repo v.Repo) (bool, error) {
	branches, err := repo.Branches()
	if err != nil {
		return false, err
	}
	for _, b := range branches {
		if b == branch {
			return true, nil
		}
	}
	return false, nil
}

// defaultBranch tries to ascertain the default branch for the given repo.
// Some repos will have multiple branches in them (e.g. Git) while others
// (e.g. Svn) will not.
func defaultBranch(repo v.Repo, home string) string {

	// Svn and Bzr use different locations (paths or entire locations)
	// for branches so we won't have a default branch.
	if repo.Vcs() == v.Svn || repo.Vcs() == v.Bzr {
		return ""
	}

	// Check the cache for a value.
	key, kerr := cp.Key(repo.Remote())
	var d cp.RepoInfo
	if kerr == nil {
		d, err := cp.RepoData(key)
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
			err := cp.SaveRepoData(key, d)
			if err == cp.ErrCacheDisabled {
				msg.Debug("Unable to cache default branch because caching is disabled")
			} else if err != nil {
				msg.Debug("Error saving %s to cache. Error: %s", repo.Remote(), err)
			}
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
			err := cp.SaveRepoData(key, d)
			if err == cp.ErrCacheDisabled {
				msg.Debug("Unable to cache default branch because caching is disabled")
			} else if err != nil {
				msg.Debug("Error saving %s to cache. Error: %s", repo.Remote(), err)
			}
		}
		return db
	}

	return ""
}

// From a local repo find out the current branch name if there is one.
// Note, this should only be used right after a fresh clone to get accurate
// information.
func findCurrentBranch(repo v.Repo) string {
	msg.Debug("Attempting to find current branch for %s", repo.Remote())
	// Svn and Bzr don't have default branches.
	if repo.Vcs() == v.Svn || repo.Vcs() == v.Bzr {
		return ""
	}

	if repo.Vcs() == v.Git || repo.Vcs() == v.Hg {
		ver, err := repo.Current()
		if err != nil {
			msg.Debug("Unable to find current branch for %s, error: %s", repo.Remote(), err)
			return ""
		}
		return ver
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
