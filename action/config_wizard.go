package action

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/semver"
	"github.com/Masterminds/vcs"
)

// ConfigWizard reads configuration from a glide.yaml file and attempts to suggest
// improvements. The wizard is interactive.
func ConfigWizard(base string) {
	_, err := gpath.Glide()
	glidefile := gpath.GlideFile
	if err != nil {
		msg.Info("Unable to find a glide.yaml file. Would you like to create one now? Yes (Y) or No (N)")
		bres := msg.PromptUntilYorN()
		if bres {
			// Guess deps
			conf := guessDeps(base, false)
			// Write YAML
			if err := conf.WriteFile(glidefile); err != nil {
				msg.Die("Could not save %s: %s", glidefile, err)
			}
		} else {
			msg.Err("Unable to find configuration file. Please create configuration information to continue.")
		}
	}

	conf := EnsureConfig()

	err = cache.Setup()
	if err != nil {
		msg.Die("Problem setting up cache: %s", err)
	}

	msg.Info("Looking for dependencies to make suggestions on")
	msg.Info("--> Scanning for dependencies not using version ranges")
	msg.Info("--> Scanning for dependencies using commit ids")
	var deps []*cfg.Dependency
	for _, dep := range conf.Imports {
		if wizardLookInto(dep) {
			deps = append(deps, dep)
		}
	}

	msg.Info("Gathering information on each dependency")
	msg.Info("--> This may take a moment. Especially on a codebase with many dependencies")
	msg.Info("--> Gathering release information for dependencies")
	msg.Info("--> Looking for dependency imports where versions are commit ids")
	for _, dep := range deps {
		wizardFindVersions(dep)
	}

	var changes int
	for _, dep := range deps {
		var remote string
		if dep.Repository != "" {
			remote = dep.Repository
		} else {
			remote = "https://" + dep.Name
		}

		// First check, ask if the tag should be used instead of the commit id for it.
		cur := cache.MemCurrent(remote)
		if cur != "" && cur != dep.Reference {
			wizardSugOnce()
			var dres bool
			asked, use, val := wizardOnce("current")
			if !use {
				dres = wizardAskCurrent(cur, dep)
			}
			if !asked {
				as := wizardRemember()
				wizardSetOnce("current", as, dres)
			}

			if asked && use {
				dres = val.(bool)
			}

			if dres {
				msg.Info("Updating %s to use the tag %s instead of commit id %s", dep.Name, cur, dep.Reference)
				dep.Reference = cur
				changes++
			}
		}

		// Second check, if no version is being used and there's a semver release ask about latest.
		memlatest := cache.MemLatest(remote)
		if dep.Reference == "" && memlatest != "" {
			wizardSugOnce()
			var dres bool
			asked, use, val := wizardOnce("latest")
			if !use {
				dres = wizardAskLatest(memlatest, dep)
			}
			if !asked {
				as := wizardRemember()
				wizardSetOnce("latest", as, dres)
			}

			if asked && use {
				dres = val.(bool)
			}

			if dres {
				msg.Info("Updating %s to use the release %s instead of no release", dep.Name, memlatest)
				dep.Reference = memlatest
				changes++
			}
		}

		// Third check, if the version is semver offer to use a range instead.
		sv, err := semver.NewVersion(dep.Reference)
		if err == nil {
			wizardSugOnce()
			var res string
			asked, use, val := wizardOnce("range")
			if !use {
				res = wizardAskRange(sv, dep)
			}
			if !asked {
				as := wizardRemember()
				wizardSetOnce("range", as, res)
			}

			if asked && use {
				res = val.(string)
			}

			if res == "m" {
				r := "^" + sv.String()
				msg.Info("Updating %s to use the range %s instead of commit id %s", dep.Name, r, dep.Reference)
				dep.Reference = r
				changes++
			} else if res == "p" {
				r := "~" + sv.String()
				msg.Info("Updating %s to use the range %s instead of commit id %s", dep.Name, r, dep.Reference)
				dep.Reference = r
				changes++
			}
		}
	}

	if changes > 0 {
		msg.Info("Configuration changes have been made. Would you like to write these")
		msg.Info("changes to your configuration file? Yes (Y) or No (N)")
		dres := msg.PromptUntilYorN()
		if dres {
			msg.Info("Writing updates to configuration file (%s)", glidefile)
			if err := conf.WriteFile(glidefile); err != nil {
				msg.Die("Could not save %s: %s", glidefile, err)
			}
			msg.Info("You can now edit the glide.yaml file.:")
			msg.Info("--> For more information on versions and ranges see https://glide.sh/docs/versions/")
			msg.Info("--> For details on additional metadata see https://glide.sh/docs/glide.yaml/")
		} else {
			msg.Warn("Change not written to configuration file")
		}
	} else {
		msg.Info("No proposed changes found. Have a nice day.")
	}
}

var wizardOnceVal = make(map[string]interface{})
var wizardOnceDo = make(map[string]bool)
var wizardOnceAsked = make(map[string]bool)

var wizardSuggeseOnce bool

func wizardSugOnce() {
	if !wizardSuggeseOnce {
		msg.Info("Here are some suggestions...")
	}
	wizardSuggeseOnce = true
}

// Returns if it's you should prompt, if not prompt if you should use stored value,
// and stored value if it has one.
func wizardOnce(name string) (bool, bool, interface{}) {
	return wizardOnceAsked[name], wizardOnceDo[name], wizardOnceVal[name]
}

func wizardSetOnce(name string, prompt bool, val interface{}) {
	wizardOnceAsked[name] = true
	wizardOnceDo[name] = prompt
	wizardOnceVal[name] = val
}

func wizardRemember() bool {
	msg.Info("Would you like to remember the previous decision and apply it to future")
	msg.Info("dependencies? Yes (Y) or No (N)")
	return msg.PromptUntilYorN()
}

func wizardAskRange(ver *semver.Version, d *cfg.Dependency) string {
	vstr := ver.String()
	msg.Info("The package %s appears to use semantic versions (http://semver.org).", d.Name)
	msg.Info("Would you like to track the latest minor or patch releases (major.minor.path)?")
	msg.Info("Tracking minor version releases would use '>= %s, < %d.0.0' ('^%s'). Tracking patch version", vstr, ver.Major()+1, vstr)
	msg.Info("releases would use '>= %s, < %d.%d.0' ('~%s'). For more information on Glide versions", vstr, ver.Major(), ver.Minor()+1, vstr)
	msg.Info("and ranges see https://glide.sh/docs/versions")
	msg.Info("Minor (M), Patch (P), or Skip Ranges (S)?")
	res, err := msg.PromptUntil([]string{"minor", "m", "patch", "p", "skip ranges", "s"})
	if err != nil {
		msg.Die("Error processing response: %s", err)
	}
	if res == "m" || res == "minor" {
		return "m"
	} else if res == "p" || res == "patch" {
		return "p"
	}

	return "s"
}

func wizardAskCurrent(cur string, d *cfg.Dependency) bool {
	msg.Info("The package %s is currently set to use the version %s.", d.Name, d.Reference)
	msg.Info("There is an equivalent semantic version (http://semver.org) release of %s. Would", cur)
	msg.Info("you like to use that instead? Yes (Y) or No (N)")
	return msg.PromptUntilYorN()
}

func wizardAskLatest(latest string, d *cfg.Dependency) bool {
	msg.Info("The package %s appears to have Semantic Version releases (http://semver.org). ", d.Name)
	msg.Info("The latestrelease is %s. You are currently not using a release. Would you like", latest)
	msg.Info("to use this release? Yes (Y) or No (N)")
	return msg.PromptUntilYorN()
}

func wizardLookInto(d *cfg.Dependency) bool {
	_, err := semver.NewConstraint(d.Reference)

	// The existing version is already a valid semver constraint so we skip suggestions.
	if err == nil {
		return false
	}

	return true
}

// Note, this really needs a simpler name.
var createGitParseVersion = regexp.MustCompile(`(?m-s)(?:tags)/(\S+)$`)

func wizardFindVersions(d *cfg.Dependency) {
	l, err := cache.Location()
	if err != nil {
		msg.Debug("Problem detecting cache location: %s", err)
		return
	}
	var remote string
	if d.Repository != "" {
		remote = d.Repository
	} else {
		remote = "https://" + d.Name
	}

	key, err := cache.Key(remote)
	if err != nil {
		msg.Debug("Problem generating cache key for %s: %s", remote, err)
		return
	}

	local := filepath.Join(l, "src", key)
	repo, err := vcs.NewRepo(remote, local)
	if err != nil {
		msg.Debug("Problem getting repo instance: %s", err)
		return
	}

	var useLocal bool
	if _, err = os.Stat(local); err == nil {
		useLocal = true
	}

	// Git endpoints allow for querying without fetching the codebase locally.
	// We try that first to avoid fetching right away. Is this premature
	// optimization?
	cc := true
	if !useLocal && repo.Vcs() == vcs.Git {
		out, err2 := exec.Command("git", "ls-remote", remote).CombinedOutput()
		if err2 == nil {
			cache.MemTouch(remote)
			cc = false
			lines := strings.Split(string(out), "\n")
			for _, i := range lines {
				ti := strings.TrimSpace(i)
				if found := createGitParseVersion.FindString(ti); found != "" {
					tg := strings.TrimPrefix(strings.TrimSuffix(found, "^{}"), "tags/")
					cache.MemPut(remote, tg)
					if d.Reference != "" && strings.HasPrefix(ti, d.Reference) {
						cache.MemSetCurrent(remote, tg)
					}
				}
			}
		}
	}

	if cc {
		cache.Lock(key)
		cache.MemTouch(remote)
		if _, err = os.Stat(local); os.IsNotExist(err) {
			repo.Get()
			branch := findCurrentBranch(repo)
			c := cache.RepoInfo{DefaultBranch: branch}
			err = cache.SaveRepoData(key, c)
			if err != nil {
				msg.Debug("Error saving cache repo details: %s", err)
			}
		} else {
			repo.Update()
		}
		tgs, err := repo.Tags()
		if err != nil {
			msg.Debug("Problem getting tags: %s", err)
		} else {
			for _, v := range tgs {
				cache.MemPut(remote, v)
			}
		}
		if d.Reference != "" && repo.IsReference(d.Reference) {
			tgs, err = repo.TagsFromCommit(d.Reference)
			if err != nil {
				msg.Debug("Problem getting tags for commit: %s", err)
			} else {
				if len(tgs) > 0 {
					for _, v := range tgs {
						if !(repo.Vcs() == vcs.Hg && v == "tip") {
							cache.MemSetCurrent(remote, v)
						}
					}
				}
			}
		}
		cache.Unlock(key)
	}
}

func findCurrentBranch(repo vcs.Repo) string {
	msg.Debug("Attempting to find current branch for %s", repo.Remote())
	// Svn and Bzr don't have default branches.
	if repo.Vcs() == vcs.Svn || repo.Vcs() == vcs.Bzr {
		return ""
	}

	if repo.Vcs() == vcs.Git || repo.Vcs() == vcs.Hg {
		ver, err := repo.Current()
		if err != nil {
			msg.Debug("Unable to find current branch for %s, error: %s", repo.Remote(), err)
			return ""
		}
		return ver
	}

	return ""
}
