package repo

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
	"github.com/codegangsta/cli"
)

// Installer provides facilities for installing the repos in a config file.
type Installer struct {

	// Force the install when certain normally stopping conditions occur.
	Force bool

	// Home is the location of cache
	Home string

	// Vendor contains the path to put the vendor packages
	Vendor string

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

// VendorPath returns the path to the location to put vendor packages
func (i *Installer) VendorPath() string {
	if i.Vendor != "" {
		return i.Vendor
	}

	vp, err := gpath.Vendor()
	if err != nil {
		return filepath.FromSlash("./vendor")
	}

	return vp
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

// Checkout reads the config file and checks out all dependencies mentioned there.
//
// This is used when initializing an empty vendor directory, or when updating a
// vendor directory based on changed config.
func (i *Installer) Checkout(conf *cfg.Config, useDev bool) error {

	dest := i.VendorPath()

	if err := ConcurrentUpdate(conf.Imports, dest, i); err != nil {
		return err
	}

	if useDev {
		return ConcurrentUpdate(conf.DevImports, dest, i)
	}

	return nil
}

// Update updates all dependencies.
//
// It begins with the dependencies in the config file, but also resolves
// transitive dependencies. The returned lockfile has all of the dependencies
// listed, but the version reconciliation has not been done.
//
// In other words, all versions in the Lockfile will be empty.
func (i *Installer) Update(conf *cfg.Config) error {
	base := "."
	vpath := i.VendorPath()

	m := &MissingPackageHandler{

		// FIXME: Where do we get the right path for this?
		destination: vpath,

		cache:       i.UseCache,
		cacheGopath: i.UseCacheGopath,
		useGopath:   i.UseGopath,
		home:        i.Home,
	}

	// Update imports
	res, err := dependency.NewResolver(base)
	if err != nil {
		msg.Die("Failed to create a resolver: %s", err)
	}
	res.Handler = m
	msg.Info("Resolving imports")
	packages, err := allPackages(conf.Imports, res)
	if err != nil {
		msg.Die("Failed to retrieve a list of dependencies: %s", err)
	}

	msg.Warn("devImports not resolved.")

	deps := depsFromPackages(packages)
	err = ConcurrentUpdate(deps, vpath, i)

	return err
}

// ConcurrentUpdate takes a list of dependencies and updates in parallel.
func ConcurrentUpdate(deps []*cfg.Dependency, cwd string, i *Installer) error {
	done := make(chan struct{}, concurrentWorkers)
	in := make(chan *cfg.Dependency, concurrentWorkers)
	var wg sync.WaitGroup
	var lock sync.Mutex
	var returnErr error

	for ii := 0; ii < concurrentWorkers; ii++ {
		go func(ch <-chan *cfg.Dependency) {
			for {
				select {
				case dep := <-ch:
					if err := VcsUpdate(dep, cwd, i); err != nil {
						msg.Warn("Update failed for %s: %s\n", dep.Name, err)

						// Capture the error while making sure the concurrent
						// operations don't step on each other.
						lock.Lock()
						if returnErr == nil {
							returnErr = err
						} else {
							returnErr = cli.NewMultiError(returnErr, err)
						}
						lock.Unlock()
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

	return returnErr
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

// MissingPackageHandler is a dependency.MissingPackageHandler.
//
// When a package is not found, this attempts to resolve and fetch.
//
// When a package is found on the GOPATH, this notifies the user.
type MissingPackageHandler struct {
	destination                   string
	home                          string
	cache, cacheGopath, useGopath bool
}

func (m *MissingPackageHandler) NotFound(pkg string) (bool, error) {
	msg.Info("Fetching %s into %s", pkg, m.destination)
	d := &cfg.Dependency{Name: pkg}
	dest := filepath.Join(m.destination, pkg)
	if err := VcsGet(d, dest, m.home, m.cache, m.cacheGopath, m.useGopath); err != nil {
		return false, err
	}
	return true, nil
}

func (m *MissingPackageHandler) OnGopath(pkg string) (bool, error) {
	// If useGopath is false, we fall back to the strategy of fetching from
	// remote.
	if !m.useGopath {
		return m.NotFound(pkg)
	}

	msg.Info("Copying package %s from the GOPATH.", pkg)
	dest := filepath.Join(m.destination, pkg)
	// Find package on Gopath
	for _, gp := range gpath.Gopaths() {
		src := filepath.Join(gp, pkg)
		// FIXME: Should probably check if src is a dir or symlink.
		if _, err := os.Stat(src); err == nil {
			if err := os.MkdirAll(dest, os.ModeDir|0755); err != nil {
				return false, err
			}
			if err := gpath.CopyDir(src, dest); err != nil {
				return false, err
			}
			return true, nil
		}
	}

	msg.Error("Could not locate %s on the GOPATH, though it was found before.", pkg)
	return false, nil
}
