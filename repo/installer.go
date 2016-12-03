package repo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/importer"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
	"github.com/Masterminds/semver"
	"github.com/Masterminds/vcs"
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

	// ResolveAllFiles enables a resolver that will examine the dependencies
	// of every file of every package, rather than only following imported
	// packages.
	ResolveAllFiles bool

	// ResolveTest sets if test dependencies should be resolved.
	ResolveTest bool

	// Updated tracks the packages that have been remotely fetched.
	Updated *UpdateTracker
}

// NewInstaller returns an Installer instance ready to use. This is the constructor.
func NewInstaller() *Installer {
	i := &Installer{}
	i.Updated = NewUpdateTracker()
	return i
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

	// Create a config setup based on the Lockfile data to process with
	// existing commands.
	newConf := &cfg.Config{}
	newConf.Name = conf.Name

	newConf.Imports = make(cfg.Dependencies, len(lock.Imports))
	for k, v := range lock.Imports {
		newConf.Imports[k] = cfg.DependencyFromLock(v)
	}

	newConf.DevImports = make(cfg.Dependencies, len(lock.DevImports))
	for k, v := range lock.DevImports {
		newConf.DevImports[k] = cfg.DependencyFromLock(v)
	}

	newConf.DeDupe()

	if len(newConf.Imports) == 0 && len(newConf.DevImports) == 0 {
		msg.Info("No dependencies found. Nothing installed.")
		return newConf, nil
	}

	msg.Info("Downloading dependencies. Please wait...")

	err := LazyConcurrentUpdate(newConf.Imports, i, newConf)
	if err != nil {
		return newConf, err
	}
	err = LazyConcurrentUpdate(newConf.DevImports, i, newConf)

	return newConf, err
}

// Checkout reads the config file and checks out all dependencies mentioned there.
//
// This is used when initializing an empty vendor directory, or when updating a
// vendor directory based on changed config.
func (i *Installer) Checkout(conf *cfg.Config) error {

	msg.Info("Downloading dependencies. Please wait...")

	if err := ConcurrentUpdate(conf.Imports, i, conf); err != nil {
		return err
	}

	if i.ResolveTest {
		return ConcurrentUpdate(conf.DevImports, i, conf)
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

	ic := newImportCache()

	m := &MissingPackageHandler{
		home:    i.Home,
		force:   i.Force,
		Config:  conf,
		Use:     ic,
		updated: i.Updated,
	}

	v := &VersionHandler{
		Use:       ic,
		Imported:  make(map[string]bool),
		Conflicts: make(map[string]bool),
		Config:    conf,
	}

	// Update imports
	res, err := dependency.NewResolver(base)
	res.ResolveTest = i.ResolveTest
	if err != nil {
		msg.Die("Failed to create a resolver: %s", err)
	}
	res.Config = conf
	res.Handler = m
	res.VersionHandler = v
	res.ResolveAllFiles = i.ResolveAllFiles
	msg.Info("Resolving imports")

	imps, timps, err := res.ResolveLocal(false)
	if err != nil {
		msg.Die("Failed to resolve local packages: %s", err)
	}
	var deps cfg.Dependencies
	var tdeps cfg.Dependencies
	for _, v := range imps {
		n := res.Stripv(v)
		if conf.HasIgnore(n) {
			continue
		}
		rt, sub := util.NormalizeName(n)
		if sub == "" {
			sub = "."
		}
		d := deps.Get(rt)
		if d == nil {
			nd := &cfg.Dependency{
				Name:        rt,
				Subpackages: []string{sub},
			}
			deps = append(deps, nd)
		} else if !d.HasSubpackage(sub) {
			d.Subpackages = append(d.Subpackages, sub)
		}
	}
	if i.ResolveTest {
		for _, v := range timps {
			n := res.Stripv(v)
			if conf.HasIgnore(n) {
				continue
			}
			rt, sub := util.NormalizeName(n)
			if sub == "" {
				sub = "."
			}
			d := deps.Get(rt)
			if d == nil {
				d = tdeps.Get(rt)
			}
			if d == nil {
				nd := &cfg.Dependency{
					Name:        rt,
					Subpackages: []string{sub},
				}
				tdeps = append(tdeps, nd)
			} else if !d.HasSubpackage(sub) {
				d.Subpackages = append(d.Subpackages, sub)
			}
		}
	}

	_, err = allPackages(deps, res, false)
	if err != nil {
		msg.Die("Failed to retrieve a list of dependencies: %s", err)
	}

	if i.ResolveTest {
		msg.Debug("Resolving test dependencies")
		_, err = allPackages(tdeps, res, true)
		if err != nil {
			msg.Die("Failed to retrieve a list of test dependencies: %s", err)
		}
	}

	msg.Info("Downloading dependencies. Please wait...")

	err = ConcurrentUpdate(conf.Imports, i, conf)
	if err != nil {
		return err
	}

	if i.ResolveTest {
		err = ConcurrentUpdate(conf.DevImports, i, conf)
		if err != nil {
			return err
		}
	}

	return nil
}

// Export from the cache to the vendor directory
func (i *Installer) Export(conf *cfg.Config) error {
	tempDir, err := ioutil.TempDir(gpath.Tmp, "glide-vendor")
	if err != nil {
		return err
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			msg.Err(err.Error())
		}
	}()

	vp := filepath.Join(tempDir, "vendor")
	err = os.MkdirAll(vp, 0755)

	msg.Info("Exporting resolved dependencies...")
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
					loc := dep.Remote()
					key, err := cache.Key(loc)
					if err != nil {
						msg.Die(err.Error())
					}
					cache.Lock(key)

					cdir := filepath.Join(cache.Location(), "src", key)
					repo, err := dep.GetRepo(cdir)
					if err != nil {
						msg.Die(err.Error())
					}
					msg.Info("--> Exporting %s", dep.Name)
					if err := repo.ExportDir(filepath.Join(vp, filepath.ToSlash(dep.Name))); err != nil {
						msg.Err("Export failed for %s: %s\n", dep.Name, err)
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
					cache.Unlock(key)
					wg.Done()
				case <-done:
					return
				}
			}
		}(in)
	}

	for _, dep := range conf.Imports {
		if !conf.HasIgnore(dep.Name) {
			err = os.MkdirAll(filepath.Join(vp, filepath.ToSlash(dep.Name)), 0755)
			if err != nil {
				lock.Lock()
				if returnErr == nil {
					returnErr = err
				} else {
					returnErr = cli.NewMultiError(returnErr, err)
				}
				lock.Unlock()
			}
			wg.Add(1)
			in <- dep
		}
	}

	if i.ResolveTest {
		for _, dep := range conf.DevImports {
			if !conf.HasIgnore(dep.Name) {
				err = os.MkdirAll(filepath.Join(vp, filepath.ToSlash(dep.Name)), 0755)
				if err != nil {
					lock.Lock()
					if returnErr == nil {
						returnErr = err
					} else {
						returnErr = cli.NewMultiError(returnErr, err)
					}
					lock.Unlock()
				}
				wg.Add(1)
				in <- dep
			}
		}
	}

	wg.Wait()

	// Close goroutines setting the version
	for ii := 0; ii < concurrentWorkers; ii++ {
		done <- struct{}{}
	}

	if returnErr != nil {
		return returnErr
	}

	msg.Info("Replacing existing vendor dependencies")

	// Check if a .git directory exists under the old vendor dir. If it does,
	// move it over to the newly-generated vendor dir - the user is probably
	// submoduling, and it's easy enough not to break their setup.
	ivg := filepath.Join(i.VendorPath(), ".git")
	_, err = os.Stat(ivg)
	if err == nil {
		msg.Info("Preserving existing vendor/.git directory")
		vpg := filepath.Join(vp, ".git")
		err = os.Rename(ivg, vpg)

		if terr, ok := err.(*os.LinkError); ok {
			err = fixcle(ivg, vpg, terr)
			if err != nil {
				msg.Warn("Failed to preserve existing vendor/.git directory")
			}
		}
	}

	err = os.RemoveAll(i.VendorPath())
	if err != nil {
		return err
	}

	err = os.Rename(vp, i.VendorPath())
	if terr, ok := err.(*os.LinkError); ok {
		return fixcle(vp, i.VendorPath(), terr)
	}

	return err

}

// fixcle is a helper function that tries to recover from cross-device rename
// errors by falling back to copying.
func fixcle(from, to string, terr *os.LinkError) error {
	// When there are different physical devices we cannot rename cross device.
	// Instead we copy.

	// syscall.EXDEV is the common name for the cross device link error
	// which has varying output text across different operating systems.
	if terr.Err == syscall.EXDEV {
		msg.Debug("Cross link err, trying manual copy: %s", terr)
		return gpath.CopyDir(from, to)
	} else if runtime.GOOS == "windows" {
		// In windows it can drop down to an operating system call that
		// returns an operating system error with a different number and
		// message. Checking for that as a fall back.
		noerr, ok := terr.Err.(syscall.Errno)
		// 0x11 (ERROR_NOT_SAME_DEVICE) is the windows error.
		// See https://msdn.microsoft.com/en-us/library/cc231199.aspx
		if ok && noerr == 0x11 {
			msg.Debug("Cross link err on Windows, trying manual copy: %s", terr)
			return gpath.CopyDir(from, to)
		}
	}

	return terr
}

// List resolves the complete dependency tree and returns a list of dependencies.
func (i *Installer) List(conf *cfg.Config) []*cfg.Dependency {
	base := "."

	ic := newImportCache()

	v := &VersionHandler{
		Use:       ic,
		Imported:  make(map[string]bool),
		Conflicts: make(map[string]bool),
		Config:    conf,
	}

	// Update imports
	res, err := dependency.NewResolver(base)
	if err != nil {
		msg.Die("Failed to create a resolver: %s", err)
	}
	res.Config = conf
	res.VersionHandler = v
	res.ResolveAllFiles = i.ResolveAllFiles

	msg.Info("Resolving imports")
	_, _, err = res.ResolveLocal(false)
	if err != nil {
		msg.Die("Failed to resolve local packages: %s", err)
	}

	_, err = allPackages(conf.Imports, res, false)
	if err != nil {
		msg.Die("Failed to retrieve a list of dependencies: %s", err)
	}

	if len(conf.DevImports) > 0 {
		msg.Warn("dev imports not resolved.")
	}

	return conf.Imports
}

// LazyConcurrentUpdate updates only deps that are not already checkout out at the right version.
//
// This is only safe when updating from a lock file.
func LazyConcurrentUpdate(deps []*cfg.Dependency, i *Installer, c *cfg.Config) error {

	newDeps := []*cfg.Dependency{}
	for _, dep := range deps {

		key, err := cache.Key(dep.Remote())
		if err != nil {
			newDeps = append(newDeps, dep)
			continue
		}
		destPath := filepath.Join(cache.Location(), "src", key)

		// Get a VCS object for this directory
		repo, err := dep.GetRepo(destPath)
		if err != nil {
			newDeps = append(newDeps, dep)
			continue
		}

		ver, err := repo.Version()
		if err != nil {
			newDeps = append(newDeps, dep)
			continue
		}
		if dep.Reference != "" {
			ci, err := repo.CommitInfo(dep.Reference)
			if err == nil && ci.Commit == dep.Reference {
				msg.Info("--> Found desired version locally %s %s!", dep.Name, dep.Reference)
				continue
			}
		}

		msg.Debug("--> Queue %s for update (%s != %s).", dep.Name, ver, dep.Reference)
		newDeps = append(newDeps, dep)
	}
	if len(newDeps) > 0 {
		return ConcurrentUpdate(newDeps, i, c)
	}

	return nil
}

// ConcurrentUpdate takes a list of dependencies and updates in parallel.
func ConcurrentUpdate(deps []*cfg.Dependency, i *Installer, c *cfg.Config) error {
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
					loc := dep.Remote()
					key, err := cache.Key(loc)
					if err != nil {
						msg.Die(err.Error())
					}
					cache.Lock(key)
					if err := VcsUpdate(dep, i.Force, i.Updated); err != nil {
						msg.Err("Update failed for %s: %s\n", dep.Name, err)
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
					cache.Unlock(key)
					wg.Done()
				case <-done:
					return
				}
			}
		}(in)
	}

	for _, dep := range deps {
		if !c.HasIgnore(dep.Name) {
			wg.Add(1)
			in <- dep
		}
	}

	wg.Wait()

	// Close goroutines setting the version
	for ii := 0; ii < concurrentWorkers; ii++ {
		done <- struct{}{}
	}

	return returnErr
}

// allPackages gets a list of all packages required to satisfy the given deps.
func allPackages(deps []*cfg.Dependency, res *dependency.Resolver, addTest bool) ([]string, error) {
	if len(deps) == 0 {
		return []string{}, nil
	}

	vdir, err := gpath.Vendor()
	if err != nil {
		return []string{}, err
	}
	vdir += string(os.PathSeparator)
	ll, err := res.ResolveAll(deps, addTest)
	if err != nil {
		return []string{}, err
	}

	for i := 0; i < len(ll); i++ {
		ll[i] = strings.TrimPrefix(ll[i], vdir)
	}
	return ll, nil
}

// MissingPackageHandler is a dependency.MissingPackageHandler.
//
// When a package is not found, this attempts to resolve and fetch.
//
// When a package is found on the GOPATH, this notifies the user.
type MissingPackageHandler struct {
	home    string
	force   bool
	Config  *cfg.Config
	Use     *importCache
	updated *UpdateTracker
}

// NotFound attempts to retrieve a package when not found in the local cache
// folder. It will attempt to get it from the remote location info.
func (m *MissingPackageHandler) NotFound(pkg string, addTest bool) (bool, error) {
	err := m.fetchToCache(pkg, addTest)
	if err != nil {
		return false, err
	}

	return true, err
}

// OnGopath will either copy a package, already found in the GOPATH, to the
// vendor/ directory or download it from the internet. This is dependent if
// useGopath on the installer is set to true to copy from the GOPATH.
func (m *MissingPackageHandler) OnGopath(pkg string, addTest bool) (bool, error) {

	err := m.fetchToCache(pkg, addTest)
	if err != nil {
		return false, err
	}

	return true, err
}

// InVendor updates a package in the vendor/ directory to make sure the latest
// is available.
func (m *MissingPackageHandler) InVendor(pkg string, addTest bool) error {
	return m.fetchToCache(pkg, addTest)
}

// PkgPath resolves the location on the filesystem where the package should be.
// This handles making sure to use the cache location.
func (m *MissingPackageHandler) PkgPath(pkg string) string {
	root, sub := util.NormalizeName(pkg)

	// For the parent applications source skip the cache.
	if root == m.Config.Name {
		pth := gpath.Basepath()
		return filepath.Join(pth, filepath.FromSlash(sub))
	}

	d := m.Config.Imports.Get(root)
	if d == nil {
		d = m.Config.DevImports.Get(root)
	}

	if d == nil {
		d, _ = m.Use.Get(root)

		if d == nil {
			d = &cfg.Dependency{Name: root}
		}
	}

	key, err := cache.Key(d.Remote())
	if err != nil {
		msg.Die("Error generating cache key for %s", d.Name)
	}

	return filepath.Join(cache.Location(), "src", key, filepath.FromSlash(sub))
}

func (m *MissingPackageHandler) fetchToCache(pkg string, addTest bool) error {
	root := util.GetRootFromPackage(pkg)
	// Skip any references to the root package.
	if root == m.Config.Name {
		return nil
	}

	d := m.Config.Imports.Get(root)
	if d == nil && addTest {
		d = m.Config.DevImports.Get(root)
	}

	// If the dependency is nil it means the Config doesn't yet know about it.
	if d == nil {
		d, _ = m.Use.Get(root)
		// We don't know about this dependency so we create a basic instance.
		if d == nil {
			d = &cfg.Dependency{Name: root}
		}

		if addTest {
			m.Config.DevImports = append(m.Config.DevImports, d)
		} else {
			m.Config.Imports = append(m.Config.Imports, d)
		}
	}

	return VcsUpdate(d, m.force, m.updated)
}

// VersionHandler handles setting the proper version in the VCS.
type VersionHandler struct {

	// If Try to use the version here if we have one. This is a cache and will
	// change over the course of setting versions.
	Use *importCache

	// Cache if importing scan has already occurred here.
	Imported map[string]bool

	Config *cfg.Config

	// There's a problem where many sub-packages have been asked to set a version
	// and you can end up with numerous conflict messages that are exactly the
	// same. We are keeping track to only display them once.
	// the parent pac
	Conflicts map[string]bool
}

// Process imports dependencies for a package
func (d *VersionHandler) Process(pkg string) (e error) {
	root := util.GetRootFromPackage(pkg)

	// Skip any references to the root package.
	if root == d.Config.Name {
		return nil
	}

	// We have not tried to import, yet.
	// Should we look in places other than the root of the project?
	if d.Imported[root] == false {
		d.Imported[root] = true
		p := d.pkgPath(root)
		f, deps, err := importer.Import(p)
		if f && err == nil {
			for _, dep := range deps {

				// The fist one wins. Would something smater than this be better?
				exists, _ := d.Use.Get(dep.Name)
				if exists == nil && (dep.Reference != "" || dep.Repository != "") {
					d.Use.Add(dep.Name, dep, root)
				}
			}
		} else if err != nil {
			msg.Err("Unable to import from %s. Err: %s", root, err)
			e = err
		}
	}

	return
}

// SetVersion sets the version for a package. If that package version is already
// set it handles the case by:
// - keeping the already set version
// - proviting messaging about the version conflict
// TODO(mattfarina): The way version setting happens can be improved. Currently not optimal.
func (d *VersionHandler) SetVersion(pkg string, addTest bool) (e error) {
	root := util.GetRootFromPackage(pkg)

	// Skip any references to the root package.
	if root == d.Config.Name {
		return nil
	}

	v := d.Config.Imports.Get(root)
	if addTest {
		if v == nil {
			v = d.Config.DevImports.Get(root)
		} else if d.Config.DevImports.Has(root) {
			// Both imports and test imports lists the same dependency.
			// There are import chains (because the import tree is resolved
			// before the test tree) that can cause this.
			tempD := d.Config.DevImports.Get(root)
			if tempD.Reference != v.Reference {
				msg.Warn("Using import %s (version %s) for test instead of testImport (version %s).", v.Name, v.Reference, tempD.Reference)
			}
			// TODO(mattfarina): Note repo difference in a warning.
		}
	}

	dep, req := d.Use.Get(root)
	if dep != nil && v != nil {
		if v.Reference == "" && dep.Reference != "" {
			v.Reference = dep.Reference
			// Clear the pin, if set, so the new version can be used.
			v.Pin = ""
			dep = v
		} else if v.Reference != "" && dep.Reference != "" && v.Reference != dep.Reference {
			dest := d.pkgPath(pkg)
			dep = determineDependency(v, dep, dest, req)
		} else {
			dep = v
		}

	} else if v != nil {
		dep = v
	} else if dep != nil {
		// We've got an imported dependency to use and don't already have a
		// record of it. Append it to the Imports.
		if addTest {
			d.Config.DevImports = append(d.Config.DevImports, dep)
		} else {
			d.Config.Imports = append(d.Config.Imports, dep)
		}
	} else {
		// If we've gotten here we don't have any depenency objects.
		r, sp := util.NormalizeName(pkg)
		dep = &cfg.Dependency{
			Name: r,
		}
		if sp != "" {
			dep.Subpackages = []string{sp}
		}
		if addTest {
			d.Config.DevImports = append(d.Config.DevImports, dep)
		} else {
			d.Config.Imports = append(d.Config.Imports, dep)
		}
	}

	err := VcsVersion(dep)
	if err != nil {
		msg.Warn("Unable to set version on %s to %s. Err: %s", root, dep.Reference, err)
		e = err
	}

	return
}

func (d *VersionHandler) pkgPath(pkg string) string {
	root, sub := util.NormalizeName(pkg)

	// For the parent applications source skip the cache.
	if root == d.Config.Name {
		pth := gpath.Basepath()
		return filepath.Join(pth, filepath.FromSlash(sub))
	}

	dep := d.Config.Imports.Get(root)
	if dep == nil {
		dep = d.Config.DevImports.Get(root)
	}

	if dep == nil {
		dep, _ = d.Use.Get(root)

		if dep == nil {
			dep = &cfg.Dependency{Name: root}
		}
	}

	key, err := cache.Key(dep.Remote())
	if err != nil {
		msg.Die("Error generating cache key for %s", dep.Name)
	}

	return filepath.Join(cache.Location(), "src", key, filepath.FromSlash(sub))
}

func determineDependency(v, dep *cfg.Dependency, dest, req string) *cfg.Dependency {
	repo, err := v.GetRepo(dest)
	if err != nil {
		singleWarn("Unable to access repo for %s\n", v.Name)
		singleInfo("Keeping %s %s", v.Name, v.Reference)
		return v
	}

	vIsRef := repo.IsReference(v.Reference)
	depIsRef := repo.IsReference(dep.Reference)

	// Both are references and they are different ones.
	if vIsRef && depIsRef {
		singleWarn("Conflict: %s rev is currently %s, but %s wants %s\n", v.Name, v.Reference, req, dep.Reference)

		displayCommitInfo(repo, v)
		displayCommitInfo(repo, dep)

		singleInfo("Keeping %s %s", v.Name, v.Reference)
		return v
	} else if vIsRef {
		// The current one is a reference and the suggestion is a SemVer constraint.
		con, err := semver.NewConstraint(dep.Reference)
		if err != nil {
			singleWarn("Version issue for %s: '%s' is neither a reference or semantic version constraint\n", dep.Name, dep.Reference)
			singleInfo("Keeping %s %s", v.Name, v.Reference)
			return v
		}

		ver, err := semver.NewVersion(v.Reference)
		if err != nil {
			// The existing version is not a semantic version.
			singleWarn("Conflict: %s version is %s, but also asked for %s\n", v.Name, v.Reference, dep.Reference)
			displayCommitInfo(repo, v)
			singleInfo("Keeping %s %s", v.Name, v.Reference)
			return v
		}

		if con.Check(ver) {
			singleInfo("Keeping %s %s because it fits constraint '%s'", v.Name, v.Reference, dep.Reference)
			return v
		}
		singleWarn("Conflict: %s version is %s but does not meet constraint '%s'\n", v.Name, v.Reference, dep.Reference)
		singleInfo("Keeping %s %s", v.Name, v.Reference)
		return v
	} else if depIsRef {

		con, err := semver.NewConstraint(v.Reference)
		if err != nil {
			singleWarn("Version issue for %s: '%s' is neither a reference or semantic version constraint\n", v.Name, v.Reference)
			singleInfo("Keeping %s %s", v.Name, v.Reference)
			return v
		}

		ver, err := semver.NewVersion(dep.Reference)
		if err != nil {
			singleWarn("Conflict: %s version is %s, but also asked for %s\n", v.Name, v.Reference, dep.Reference)
			displayCommitInfo(repo, dep)
			singleInfo("Keeping %s %s", v.Name, v.Reference)
			return v
		}

		if con.Check(ver) {
			v.Reference = dep.Reference
			singleInfo("Using %s %s because it fits constraint '%s'", v.Name, v.Reference, v.Reference)
			return v
		}
		singleWarn("Conflict: %s semantic version constraint is %s but '%s' does not meet the constraint\n", v.Name, v.Reference, v.Reference)
		singleInfo("Keeping %s %s", v.Name, v.Reference)
		return v
	}
	// Neither is a vcs reference and both could be semantic version
	// constraints that are different.

	_, err = semver.NewConstraint(dep.Reference)
	if err != nil {
		// dd.Reference is not a reference or a valid constraint.
		singleWarn("Version %s %s is not a reference or valid semantic version constraint\n", dep.Name, dep.Reference)
		singleInfo("Keeping %s %s", v.Name, v.Reference)
		return v
	}

	_, err = semver.NewConstraint(v.Reference)
	if err != nil {
		// existing.Reference is not a reference or a valid constraint.
		// We really should never end up here.
		singleWarn("Version %s %s is not a reference or valid semantic version constraint\n", v.Name, v.Reference)

		v.Reference = dep.Reference
		v.Pin = ""
		singleInfo("Using %s %s because it is a valid version", v.Name, v.Reference)
		return v
	}

	// Both versions are constraints. Try to merge them.
	// If either comparison has an || skip merging. That's complicated.
	ddor := strings.Index(dep.Reference, "||")
	eor := strings.Index(v.Reference, "||")
	if ddor == -1 && eor == -1 {
		// Add the comparisons together.
		newRef := v.Reference + ", " + dep.Reference
		v.Reference = newRef
		v.Pin = ""
		singleInfo("Combining %s semantic version constraints %s and %s", v.Name, v.Reference, dep.Reference)
		return v
	}
	singleWarn("Conflict: %s version is %s, but also asked for %s\n", v.Name, v.Reference, dep.Reference)
	singleInfo("Keeping %s %s", v.Name, v.Reference)
	return v
}

var warningMessage = make(map[string]bool)
var infoMessage = make(map[string]bool)

func singleWarn(ft string, v ...interface{}) {
	m := fmt.Sprintf(ft, v...)
	_, f := warningMessage[m]
	if !f {
		msg.Warn(m)
		warningMessage[m] = true
	}
}

func singleInfo(ft string, v ...interface{}) {
	m := fmt.Sprintf(ft, v...)
	_, f := infoMessage[m]
	if !f {
		msg.Info(m)
		infoMessage[m] = true
	}
}

type importCache struct {
	cache map[string]*cfg.Dependency
	from  map[string]string
}

func newImportCache() *importCache {
	return &importCache{
		cache: make(map[string]*cfg.Dependency),
		from:  make(map[string]string),
	}
}

func (i *importCache) Get(name string) (*cfg.Dependency, string) {
	d, f := i.cache[name]
	if f {
		return d, i.from[name]
	}

	return nil, ""
}

func (i *importCache) Add(name string, dep *cfg.Dependency, root string) {
	i.cache[name] = dep
	i.from[name] = root
}

var displayCommitInfoPrefix = msg.Default.Color(msg.Green, "[INFO] ")
var displayCommitInfoTemplate = "%s reference %s:\n" +
	displayCommitInfoPrefix + "- author: %s\n" +
	displayCommitInfoPrefix + "- commit date: %s\n" +
	displayCommitInfoPrefix + "- subject (first line): %s\n"

func displayCommitInfo(repo vcs.Repo, dep *cfg.Dependency) {
	c, err := repo.CommitInfo(dep.Reference)
	ref := dep.Reference

	if err == nil {
		tgs, err2 := repo.TagsFromCommit(c.Commit)
		if err2 == nil && len(tgs) > 0 {
			if tgs[0] != dep.Reference {
				ref = ref + " (" + tgs[0] + ")"
			}
		}
		singleInfo(displayCommitInfoTemplate, dep.Name, ref, c.Author, c.Date.Format(time.RFC1123Z), commitSubjectFirstLine(c.Message))
	}
}

func commitSubjectFirstLine(sub string) string {
	lines := strings.Split(sub, "\n")
	if len(lines) <= 1 {
		return sub
	}

	return lines[0]
}
