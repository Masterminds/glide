package action

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/gb"
	"github.com/Masterminds/glide/godep"
	"github.com/Masterminds/glide/gom"
	"github.com/Masterminds/glide/gpm"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
	"github.com/Masterminds/vcs"
)

// Create creates/initializes a new Glide repository.
//
// This will fail if a glide.yaml already exists.
//
// By default, this will scan the present source code directory for dependencies.
//
// If skipImport is set to true, this will not attempt to import from an existing
// GPM, Godep, or GB project if one should exist. However, it will still attempt
// to read the local source to determine required packages.
func Create(base string, skipImport bool) {
	glidefile := gpath.GlideFile
	// Guard against overwrites.
	guardYAML(glidefile)

	// Guess deps
	conf := guessDeps(base, skipImport)
	// Write YAML
	msg.Info("Writing glide.yaml file")
	if err := conf.WriteFile(glidefile); err != nil {
		msg.Die("Could not save %s: %s", glidefile, err)
	}
	msg.Info("You can now edit the glide.yaml file. Consider:")
	msg.Info("--> Using versions and ranges. See https://glide.sh/docs/versions/")
	msg.Info("--> Adding additional metadata. See https://glide.sh/docs/glide.yaml/")
}

// guardYAML fails if the given file already exists.
//
// This prevents an important file from being overwritten.
func guardYAML(filename string) {
	if _, err := os.Stat(filename); err == nil {
		msg.Die("Cowardly refusing to overwrite existing YAML.")
	}
}

// guessDeps attempts to resolve all of the dependencies for a given project.
//
// base is the directory to start with.
// skipImport will skip running the automatic imports.
//
// FIXME: This function is likely a one-off that has a more standard alternative.
// It's also long and could use a refactor.
func guessDeps(base string, skipImport bool) *cfg.Config {
	buildContext, err := util.GetBuildContext()
	if err != nil {
		msg.Die("Failed to build an import context: %s", err)
	}
	name := buildContext.PackageName(base)

	msg.Info("Generating a YAML configuration file and guessing the dependencies")

	config := new(cfg.Config)

	// Get the name of the top level package
	config.Name = name

	// Import by looking at other package managers and looking over the
	// entire directory structure.

	// Attempt to import from other package managers.
	if !skipImport {
		guessImportDeps(base, config)
	}

	importLen := len(config.Imports)
	if importLen > 0 {
		msg.Info("Scanning code to look for additional dependencies")
	} else if !skipImport {
		msg.Info("No dependencies found to import")
		msg.Info("Scanning code to look for dependencies")
	} else {
		msg.Info("Scanning code to look for dependencies")
	}

	// Resolve dependencies by looking at the tree.
	r, err := dependency.NewResolver(base)
	if err != nil {
		msg.Die("Error creating a dependency resolver: %s", err)
	}

	h := &dependency.DefaultMissingPackageHandler{Missing: []string{}, Gopath: []string{}}
	r.Handler = h

	sortable, err := r.ResolveLocal(false)
	if err != nil {
		msg.Die("Error resolving local dependencies: %s", err)
	}

	sort.Strings(sortable)

	vpath := r.VendorDir
	if !strings.HasSuffix(vpath, "/") {
		vpath = vpath + string(os.PathSeparator)
	}

	var count int
	for _, pa := range sortable {
		n := strings.TrimPrefix(pa, vpath)
		root, subpkg := util.NormalizeName(n)

		if !config.HasDependency(root) {
			count++
			msg.Info("--> Found reference to %s\n", n)
			d := &cfg.Dependency{
				Name: root,
			}
			if len(subpkg) > 0 {
				d.Subpackages = []string{subpkg}
			}
			config.Imports = append(config.Imports, d)
		} else {
			if len(subpkg) > 0 {
				subpkg = strings.TrimPrefix(subpkg, "/")
				d := config.Imports.Get(root)
				if !d.HasSubpackage(subpkg) {
					count++
					msg.Info("--> Adding sub-package %s to %s\n", subpkg, root)
					d.Subpackages = append(d.Subpackages, subpkg)
				}
			}
		}
	}

	if count == 0 && importLen > 0 {
		msg.Info("--> Scanning found no additional dependencies to import")
	} else if count == 0 {
		msg.Info("--> Scanning found no dependencies to import")
	}

	return config
}

func guessImportDeps(base string, config *cfg.Config) {
	msg.Info("Attempting to import from other package managers (use --skip-import to skip)")
	deps := []*cfg.Dependency{}
	absBase, err := filepath.Abs(base)
	if err != nil {
		msg.Die("Failed to resolve location of %s: %s", base, err)
	}

	if d, ok := guessImportGodep(absBase); ok {
		msg.Info("Importing Godep configuration")
		msg.Warn("--> Godep uses commit id versions. Consider using Semantic Versions with Glide")
		deps = d
	} else if d, ok := guessImportGPM(absBase); ok {
		msg.Info("Importing GPM configuration")
		deps = d
	} else if d, ok := guessImportGB(absBase); ok {
		msg.Info("Importing GB configuration")
		deps = d
	} else if d, ok := guessImportGom(absBase); ok {
		msg.Info("Importing GB configuration")
		deps = d
	}

	if len(deps) > 0 {
		msg.Info("--> Attempting to detect versions from commit ids")
	}

	var wg sync.WaitGroup

	for _, i := range deps {
		wg.Add(1)
		go func(dep *cfg.Dependency) {
			var remote string
			if dep.Repository != "" {
				remote = dep.Repository
			} else {
				remote = "https://" + dep.Name
			}
			ver := createGuessVersion(remote, dep.Reference)
			if ver != dep.Reference {
				msg.Info("--> Found imported reference to %s at version %s", dep.Name, ver)
				dep.Reference = ver
			} else {
				msg.Info("--> Found imported reference to %s", dep.Name)
			}

			wg.Done()
		}(i)

		config.Imports = append(config.Imports, i)
	}

	wg.Wait()
}

func guessImportGodep(dir string) ([]*cfg.Dependency, bool) {
	d, err := godep.Parse(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}

func guessImportGPM(dir string) ([]*cfg.Dependency, bool) {
	d, err := gpm.Parse(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}

func guessImportGB(dir string) ([]*cfg.Dependency, bool) {
	d, err := gb.Parse(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}

func guessImportGom(dir string) ([]*cfg.Dependency, bool) {
	d, err := gom.Parse(dir)
	if err != nil || len(d) == 0 {
		return []*cfg.Dependency{}, false
	}

	return d, true
}

// Note, this really needs a simpler name.
var createGitParseVersion = regexp.MustCompile(`(?m-s)(?:tags)/(\S+)$`)

func createGuessVersion(remote, id string) string {
	err := cache.Setup()
	if err != nil {
		msg.Debug("Problem setting up cache: %s", err)
	}
	l, err := cache.Location()
	if err != nil {
		msg.Debug("Problem detecting cache location: %s", err)
	}
	key, err := cache.Key(remote)
	if err != nil {
		msg.Debug("Problem generating cache key for %s: %s", remote, err)
	}

	local := filepath.Join(l, "src", key)
	repo, err := vcs.NewRepo(remote, local)
	if err != nil {
		msg.Debug("Problem getting repo instance: %s", err)
	}

	// Git endpoints allow for querying without fetching the codebase locally.
	// We try that first to avoid fetching right away. Is this premature
	// optimization?
	cc := true
	if repo.Vcs() == vcs.Git {
		out, err := exec.Command("git", "ls-remote", remote).CombinedOutput()
		if err == nil {
			cc = false
			lines := strings.Split(string(out), "\n")

			// TODO(mattfarina): Detect if the found version is semver and use
			// that one instead of the first found.
			for _, i := range lines {
				ti := strings.TrimSpace(i)
				if strings.HasPrefix(ti, id) {
					if found := createGitParseVersion.FindString(ti); found != "" {
						return strings.TrimPrefix(strings.TrimSuffix(found, "^{}"), "tags/")
					}
				}
			}
		}
	}

	if cc {
		cache.Lock(key)
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

		tgs, err := repo.TagsFromCommit(id)
		if err != nil {
			msg.Debug("Problem getting tags for commit: %s", err)
		}
		cache.Unlock(key)
		if len(tgs) > 0 {
			return tgs[0]
		}
	}

	return id
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
