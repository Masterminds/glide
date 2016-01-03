package repo

import (
	"os"
	"strings"
	"sync"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
)

// Installer provides facilities for installing the repos in a config file.
type Installer struct {

	// Force the install when certain normally stopping conditions occur.
	Force bool

	// Home is the location of cache
	Home string

	// Use a cache
	UseCache bool
	// Use Gopath to cache
	UseCacheGopath bool
	// Use Gopath as a source to read from
	UseGopath bool

	// UpdateVendored instructs the environment to update in a way that is friendly
	// to packages that have been "vendored in" (e.g. are copies of source, not repos)
	UpdateVendored bool

	// DeleteUnused deletes packages that are unused, but found in the vendor dir.
	DeleteUnused bool
}

// Install installs the dependencies from a Lockfile.
func (i *Installer) Install(lock *cfg.Lockfile, conf *cfg.Config) (*cfg.Config, error) {

	cwd, err := gpath.Vendor()
	if err != nil {
		return conf, err
	}

	// Create a config setup based on the Lockfile data to process with
	// existing commands.
	newConf := &cfg.Config{}
	newConf.Name = conf.Name

	newConf.Imports = make(cfg.Dependencies, len(lock.Imports))
	for k, v := range lock.Imports {
		newConf.Imports[k] = &cfg.Dependency{
			Name:        v.Name,
			Reference:   v.Version,
			Repository:  v.Repository,
			VcsType:     v.VcsType,
			Subpackages: v.Subpackages,
			Arch:        v.Arch,
			Os:          v.Os,
		}
	}

	newConf.DevImports = make(cfg.Dependencies, len(lock.DevImports))
	for k, v := range lock.DevImports {
		newConf.DevImports[k] = &cfg.Dependency{
			Name:        v.Name,
			Reference:   v.Version,
			Repository:  v.Repository,
			VcsType:     v.VcsType,
			Subpackages: v.Subpackages,
			Arch:        v.Arch,
			Os:          v.Os,
		}
	}

	newConf.DeDupe()

	if len(newConf.Imports) == 0 {
		msg.Info("No dependencies found. Nothing installed.\n")
		return newConf, nil
	}

	ConcurrentUpdate(newConf.Imports, cwd, i)
	ConcurrentUpdate(newConf.DevImports, cwd, i)
	return newConf, nil
}

// Update updates all dependencies.
//
// It begins with the dependencies in the config file, but also resolves
// transitive dependencies. The returned lockfile has all of the dependencies
// listed, but the version reconciliation has not been done.
//
// In other words, all versions in the Lockfile will be empty.
func (i *Installer) Update(conf *cfg.Config) (*cfg.Lockfile, error) {
	base := "."

	// Update imports
	res, err := dependency.NewResolver(base)
	if err != nil {
		msg.Die("Failed to create a resolver: %s", err)
	}
	msg.Info("Resolving imports")
	packages, err := allPackages(conf.Imports, res)
	if err != nil {
		msg.Die("Failed to retrieve a list of dependencies: %s", err)
	}

	msg.Warn("devImports not resolved.")

	deps := depsFromPackages(packages)
	ConcurrentUpdate(deps, base, i)

	hash, err := conf.Hash()
	if err != nil {
		return nil, err
	}
	return cfg.NewLockfile(deps, hash), nil
}

// ConcurrentUpdate takes a list of dependencies and updates in parallel.
func ConcurrentUpdate(deps []*cfg.Dependency, cwd string, i *Installer) {
	done := make(chan struct{}, concurrentWorkers)
	in := make(chan *cfg.Dependency, concurrentWorkers)
	var wg sync.WaitGroup

	for ii := 0; ii < concurrentWorkers; ii++ {
		go func(ch <-chan *cfg.Dependency) {
			for {
				select {
				case dep := <-ch:
					if err := VcsUpdate(dep, cwd, i); err != nil {
						msg.Warn("Update failed for %s: %s\n", dep.Name, err)
					}
					wg.Done()
				case <-done:
					return
				}
			}
		}(in)
	}

	for _, dep := range deps {
		wg.Add(1)
		in <- dep
	}

	wg.Wait()

	// Close goroutines setting the version
	for ii := 0; ii < concurrentWorkers; ii++ {
		done <- struct{}{}
	}
}

// allPackages gets a list of all packages required to satisfy the given deps.
func allPackages(deps []*cfg.Dependency, res *dependency.Resolver) ([]string, error) {
	if len(deps) == 0 {
		return []string{}, nil
	}

	vdir, err := gpath.Vendor()
	if err != nil {
		return []string{}, err
	}
	vdir += string(os.PathSeparator)
	ll, err := res.ResolveAll(deps)
	if err != nil {
		return []string{}, err
	}

	for i := 0; i < len(ll); i++ {
		ll[i] = strings.TrimPrefix(ll[i], vdir)
	}
	return ll, nil
}

/* unused
func reposFromPackages(pkgs []string) []string {
	// Make sure we don't have to resize this.
	seen := make(map[string]bool, len(pkgs))

	// Order is important.
	repos := []string{}

	for _, p := range pkgs {
		rr, _ := util.NormalizeName(p)
		if !seen[rr] {
			seen[rr] = true
			repos = append(repos, rr)
		}
	}
	return repos
}
*/

func depsFromPackages(pkgs []string) []*cfg.Dependency {
	// Make sure we don't have to resize this.
	seen := make(map[string]*cfg.Dependency, len(pkgs))

	// Order is important.
	deps := []*cfg.Dependency{}

	for _, p := range pkgs {
		rr, sp := util.NormalizeName(p)
		if _, ok := seen[rr]; !ok {
			subpkg := []string{}
			if sp != "" {
				subpkg = append(subpkg, sp)
			}

			dd := &cfg.Dependency{
				Name:        rr,
				Subpackages: subpkg,
			}

			deps = append(deps, dd)
			seen[rr] = dd
		} else if sp != "" {
			seen[rr].Subpackages = append(seen[rr].Subpackages, sp)
		}
	}
	return deps
}
