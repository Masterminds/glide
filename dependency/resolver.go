package dependency

import (
	"container/list"
	"runtime"
	//"go/build"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
)

// MissingPackageHandler handles the case where a package is missing during scanning.
//
// It returns true if the package can be passed to the resolver, false otherwise.
// False may be returned even if error is nil.
type MissingPackageHandler interface {
	// NotFound is called when the Resolver fails to find a package with the given name.
	//
	// NotFound returns true when the resolver should attempt to re-resole the
	// dependency (e.g. when NotFound has gone and fetched the missing package).
	//
	// When NotFound returns false, the Resolver does not try to do any additional
	// work on the missing package.
	//
	// NotFound only returns errors when it fails to perform its internal goals.
	// When it returns false with no error, this indicates that the handler did
	// its job, but the resolver should not do any additional work on the
	// package.
	NotFound(pkg string, addTest bool) (bool, error)

	// OnGopath is called when the Resolver finds a dependency, but it's only on GOPATH.
	//
	// OnGopath provides an opportunity to copy, move, warn, or ignore cases like this.
	//
	// OnGopath returns true when the resolver should attempt to re-resolve the
	// dependency (e.g. when the dependency is copied to a new location).
	//
	// When OnGopath returns false, the Resolver does not try to do any additional
	// work on the package.
	//
	// An error indicates that OnGopath cannot complete its intended operation.
	// Not all false results are errors.
	OnGopath(pkg string, addTest bool) (bool, error)

	// InVendor is called when the Resolver finds a dependency in the vendor/ directory.
	//
	// This can be used update a project found in the vendor/ folder.
	InVendor(pkg string, addTest bool) error
}

// DefaultMissingPackageHandler is the default handler for missing packages.
//
// When asked to handle a missing package, it will report the miss as a warning,
// and then store the package in the Missing slice for later access.
type DefaultMissingPackageHandler struct {
	Missing []string
	Gopath  []string
}

// NotFound prints a warning and then stores the package name in Missing.
//
// It never returns an error, and it always returns false.
func (d *DefaultMissingPackageHandler) NotFound(pkg string, addTest bool) (bool, error) {
	msg.Warn("Package %s is not installed", pkg)
	d.Missing = append(d.Missing, pkg)
	return false, nil
}

// OnGopath is run when a package is missing from vendor/ but found in the GOPATH
func (d *DefaultMissingPackageHandler) OnGopath(pkg string, addTest bool) (bool, error) {
	msg.Warn("Package %s is only on GOPATH.", pkg)
	d.Gopath = append(d.Gopath, pkg)
	return false, nil
}

// InVendor is run when a package is found in the vendor/ folder
func (d *DefaultMissingPackageHandler) InVendor(pkg string, addTest bool) error {
	msg.Info("Package %s found in vendor/ folder", pkg)
	return nil
}

// VersionHandler sets the version for a package when found while scanning.
//
// When a package if found it needs to be on the correct version before
// scanning its contents to be sure to pick up the right elements for that
// version.
type VersionHandler interface {

	// Process provides an opportunity to process the codebase for version setting.
	Process(pkg string) error

	// SetVersion sets the version for a package. An error is returned if there
	// was a problem setting the version.
	SetVersion(pkg string, testDep bool) error
}

// DefaultVersionHandler is the default handler for setting the version.
//
// The default handler leaves the current version and skips setting a version.
// For a handler that alters the version see the handler included in the repo
// package as part of the installer.
type DefaultVersionHandler struct{}

// Process a package to aide in version setting.
func (d *DefaultVersionHandler) Process(pkg string) error {
	return nil
}

// SetVersion here sends a message when a package is found noting that it
// did not set the version.
func (d *DefaultVersionHandler) SetVersion(pkg string, testDep bool) error {
	msg.Warn("Version not set for package %s", pkg)
	return nil
}

// Resolver resolves a dependency tree.
//
// It operates in two modes:
// - local resolution (ResolveLocal) determines the dependencies of the local project.
// - vendor resolving (Resolve, ResolveAll) determines the dependencies of vendored
//   projects.
//
// Local resolution is for guessing initial dependencies. Vendor resolution is
// for determining vendored dependencies.
type Resolver struct {
	Handler        MissingPackageHandler
	VersionHandler VersionHandler
	VendorDir      string
	BuildContext   *util.BuildCtxt
	Config         *cfg.Config

	// ResolveAllFiles toggles deep scanning.
	// If this is true, resolve by scanning all files, not by walking the
	// import tree.
	ResolveAllFiles bool

	// ResolveTest sets if test dependencies should be resolved.
	ResolveTest bool

	// Items already in the queue.
	alreadyQ map[string]bool

	// Attempts to scan that had unrecoverable error.
	hadError map[string]bool

	basedir string
	seen    map[string]bool

	// findCache caches hits from Find. This reduces the number of filesystem
	// touches that have to be done for dependency resolution.
	findCache map[string]*PkgInfo
}

// NewResolver returns a new Resolver initialized with the DefaultMissingPackageHandler.
//
// This will return an error if the given path does not meet the basic criteria
// for a Go source project. For example, basedir must have a vendor subdirectory.
//
// The BuildContext uses the "go/build".Default to resolve dependencies.
func NewResolver(basedir string) (*Resolver, error) {

	var err error
	basedir, err = filepath.Abs(basedir)
	if err != nil {
		return nil, err
	}

	basedir, err = checkForBasedirSymlink(basedir)

	if err != nil {
		return nil, err
	}

	vdir := filepath.Join(basedir, "vendor")

	buildContext, err := util.GetBuildContext()
	if err != nil {
		return nil, err
	}

	r := &Resolver{
		Handler:        &DefaultMissingPackageHandler{Missing: []string{}, Gopath: []string{}},
		VersionHandler: &DefaultVersionHandler{},
		basedir:        basedir,
		VendorDir:      vdir,
		BuildContext:   buildContext,
		seen:           map[string]bool{},
		alreadyQ:       map[string]bool{},
		hadError:       map[string]bool{},
		findCache:      map[string]*PkgInfo{},

		// The config instance here should really be replaced with a real one.
		Config: &cfg.Config{},
	}

	// TODO: Make sure the build context is correctly set up. Especially in
	// regards to GOROOT, which is not always set.

	return r, nil
}

// Resolve takes a package name and returns all of the imported package names.
//
// If a package is not found, this calls the Fetcher. If the Fetcher returns
// true, it will re-try traversing that package for dependencies. Otherwise it
// will add that package to the deps array and continue on without trying it.
// And if the Fetcher returns an error, this will stop resolution and return
// the error.
//
// If basepath is set to $GOPATH, this will start from that package's root there.
// If basepath is set to a project's vendor path, the scanning will begin from
// there.
func (r *Resolver) Resolve(pkg, basepath string) ([]string, error) {
	target := filepath.Join(basepath, filepath.FromSlash(pkg))
	//msg.Debug("Scanning %s", target)
	l := list.New()
	l.PushBack(target)

	// In this mode, walk the entire tree.
	if r.ResolveAllFiles {
		return r.resolveList(l, false, false)
	}
	return r.resolveImports(l, false, false)
}

// dirHasPrefix tests whether the directory dir begins with prefix.
func dirHasPrefix(dir, prefix string) bool {
	if runtime.GOOS != "windows" {
		return strings.HasPrefix(dir, prefix)
	}
	return len(dir) >= len(prefix) && strings.EqualFold(dir[:len(prefix)], prefix)
}

// ResolveLocal resolves dependencies for the current project.
//
// This begins with the project, builds up a list of external dependencies.
//
// If the deep flag is set to true, this will then resolve all of the dependencies
// of the dependencies it has found. If not, it will return just the packages that
// the base project relies upon.
func (r *Resolver) ResolveLocal(deep bool) ([]string, []string, error) {
	// We build a list of local source to walk, then send this list
	// to resolveList.
	msg.Debug("Resolving local dependencies")
	l := list.New()
	tl := list.New()
	alreadySeen := map[string]bool{}
	talreadySeen := map[string]bool{}
	err := filepath.Walk(r.basedir, func(path string, fi os.FileInfo, err error) error {
		if err != nil && err != filepath.SkipDir {
			return err
		}
		pt := strings.TrimPrefix(path, r.basedir+string(os.PathSeparator))
		pt = strings.TrimSuffix(pt, string(os.PathSeparator))
		if r.Config.HasExclude(pt) {
			msg.Debug("Excluding %s", pt)
			return filepath.SkipDir
		}
		if !fi.IsDir() {
			return nil
		}
		if !srcDir(fi) {
			return filepath.SkipDir
		}

		// Scan for dependencies, and anything that's not part of the local
		// package gets added to the scan list.
		var imps []string
		var testImps []string
		p, err := r.BuildContext.ImportDir(path, 0)
		if err != nil {
			if strings.HasPrefix(err.Error(), "no buildable Go source") {
				return nil
			} else if strings.HasPrefix(err.Error(), "found packages ") {
				// If we got here it's because a package and multiple packages
				// declared. This is often because of an example with a package
				// or main but +build ignore as a build tag. In that case we
				// try to brute force the packages with a slower scan.
				imps, testImps, err = IterativeScan(path)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			imps = p.Imports
			testImps = p.TestImports
		}

		// We are only looking for dependencies in vendor. No root, cgo, etc.
		for _, imp := range imps {
			if r.Config.HasIgnore(imp) {
				continue
			}
			if alreadySeen[imp] {
				continue
			}
			alreadySeen[imp] = true
			info := r.FindPkg(imp)
			switch info.Loc {
			case LocUnknown, LocVendor:
				l.PushBack(filepath.Join(r.VendorDir, filepath.FromSlash(imp))) // Do we need a path on this?
			case LocGopath:
				if !dirHasPrefix(info.Path, r.basedir) {
					// FIXME: This is a package outside of the project we're
					// scanning. It should really be on vendor. But we don't
					// want it to reference GOPATH. We want it to be detected
					// and moved.
					l.PushBack(filepath.Join(r.VendorDir, filepath.FromSlash(imp)))
				}
			case LocRelative:
				if strings.HasPrefix(imp, "./"+gpath.VendorDir) {
					msg.Warn("Go package resolving will resolve %s without the ./%s/ prefix", imp, gpath.VendorDir)
				}
			}
		}

		if r.ResolveTest {
			for _, imp := range testImps {
				if talreadySeen[imp] {
					continue
				}
				talreadySeen[imp] = true
				info := r.FindPkg(imp)
				switch info.Loc {
				case LocUnknown, LocVendor:
					tl.PushBack(filepath.Join(r.VendorDir, filepath.FromSlash(imp))) // Do we need a path on this?
				case LocGopath:
					if !dirHasPrefix(info.Path, r.basedir) {
						// FIXME: This is a package outside of the project we're
						// scanning. It should really be on vendor. But we don't
						// want it to reference GOPATH. We want it to be detected
						// and moved.
						tl.PushBack(filepath.Join(r.VendorDir, filepath.FromSlash(imp)))
					}
				case LocRelative:
					if strings.HasPrefix(imp, "./"+gpath.VendorDir) {
						msg.Warn("Go package resolving will resolve %s without the ./%s/ prefix", imp, gpath.VendorDir)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		msg.Err("Failed to build an initial list of packages to scan: %s", err)
		return []string{}, []string{}, err
	}

	if deep {
		if r.ResolveAllFiles {
			re, err := r.resolveList(l, false, false)
			if err != nil {
				return []string{}, []string{}, err
			}
			tre, err := r.resolveList(l, false, true)
			return re, tre, err
		}
		re, err := r.resolveImports(l, false, false)
		if err != nil {
			return []string{}, []string{}, err
		}
		tre, err := r.resolveImports(tl, true, true)
		return re, tre, err
	}

	// If we're not doing a deep scan, we just convert the list into an
	// array and return.
	res := make([]string, 0, l.Len())
	for e := l.Front(); e != nil; e = e.Next() {
		res = append(res, e.Value.(string))
	}
	tres := make([]string, 0, l.Len())
	if r.ResolveTest {
		for e := tl.Front(); e != nil; e = e.Next() {
			tres = append(tres, e.Value.(string))
		}
	}

	return res, tres, nil
}

// ResolveAll takes a list of packages and returns an inclusive list of all
// vendored dependencies.
//
// While this will scan all of the source code it can find, it will only return
// packages that were either explicitly passed in as deps, or were explicitly
// imported by the code.
//
// Packages that are either CGO or on GOROOT are ignored. Packages that are
// on GOPATH, but not vendored currently generate a warning.
//
// If one of the passed in packages does not exist in the vendor directory,
// an error is returned.
func (r *Resolver) ResolveAll(deps []*cfg.Dependency, addTest bool) ([]string, error) {

	queue := sliceToQueue(deps, r.VendorDir)

	if r.ResolveAllFiles {
		return r.resolveList(queue, false, addTest)
	}
	return r.resolveImports(queue, false, addTest)
}

// Stripv strips the vendor/ prefix from vendored packages.
func (r *Resolver) Stripv(str string) string {
	return strings.TrimPrefix(str, r.VendorDir+string(os.PathSeparator))
}

// vpath adds an absolute vendor path.
func (r *Resolver) vpath(str string) string {
	return filepath.Join(r.basedir, "vendor", str)
}

// resolveImports takes a list of existing packages and resolves their imports.
//
// It returns a list of all of the packages that it can determine are required
// for the given code to function.
//
// The expectation is that each item in the queue is an absolute path to a
// vendored package. This attempts to read that package, and then find
// its referenced packages. Those packages are then added to the list
// to be scanned next.
//
// The resolver's handler is used in the cases where a package cannot be
// located.
//
// testDeps specifies if the test dependencies should be resolved and addTest
// specifies if the dependencies should be added to the Config.DevImports. This
// is important because we may resolve normal dependencies of test deps and add
// them to the DevImports list.
func (r *Resolver) resolveImports(queue *list.List, testDeps, addTest bool) ([]string, error) {
	msg.Debug("Resolving import path")

	// When test deps passed in but not resolving return empty.
	if (testDeps || addTest) && !r.ResolveTest {
		return []string{}, nil
	}

	for e := queue.Front(); e != nil; e = e.Next() {
		vdep := e.Value.(string)
		dep := r.Stripv(vdep)
		// Check if marked in the Q and then explicitly mark it. We want to know
		// if it had previously been marked and ensure it for the future.

		_, foundQ := r.alreadyQ[dep]
		r.alreadyQ[dep] = true

		// If we've already encountered an error processing this dependency
		// skip it.
		_, foundErr := r.hadError[dep]
		if foundErr {
			continue
		}

		// Skip ignored packages
		if r.Config.HasIgnore(dep) {
			msg.Debug("Ignoring: %s", dep)
			continue
		}
		r.VersionHandler.Process(dep)
		// Here, we want to import the package and see what imports it has.
		msg.Debug("Trying to open %s", vdep)
		var imps []string
		pkg, err := r.BuildContext.ImportDir(vdep, 0)
		if err != nil && strings.HasPrefix(err.Error(), "found packages ") {
			// If we got here it's because a package and multiple packages
			// declared. This is often because of an example with a package
			// or main but +build ignore as a build tag. In that case we
			// try to brute force the packages with a slower scan.
			msg.Debug("Using Iterative Scanning for %s", dep)
			if testDeps {
				_, imps, err = IterativeScan(vdep)
			} else {
				imps, _, err = IterativeScan(vdep)
			}

			if err != nil {
				msg.Err("Iterative scanning error %s: %s", dep, err)
				continue
			}
		} else if err != nil {
			msg.Debug("ImportDir error on %s: %s", vdep, err)
			if strings.HasPrefix(err.Error(), "no buildable Go source") {
				msg.Debug("No subpackages declared. Skipping %s.", dep)
				continue
			} else if os.IsNotExist(err) && !foundErr && !foundQ {
				// If the location doesn't exist, there hasn't already been an
				// error, it's not already been in the Q then try to fetch it.
				// When there's an error or it's already in the Q (it should be
				// fetched if it's marked in r.alreadyQ) we skip to make sure
				// not to get stuck in a recursion.

				// If the location doesn't exist try to fetch it.
				if ok, err2 := r.Handler.NotFound(dep, addTest); ok {
					r.alreadyQ[dep] = true

					// By adding to the queue it will get reprocessed now that
					// it exists.
					queue.PushBack(r.vpath(dep))
					r.VersionHandler.SetVersion(dep, addTest)
				} else if err2 != nil {
					r.hadError[dep] = true
					msg.Err("Error looking for %s: %s", dep, err2)
				} else {
					r.hadError[dep] = true
					// TODO (mpb): Should we toss this into a Handler to
					// see if this is on GOPATH and copy it?
					msg.Info("Not found in vendor/: %s (1)", dep)
				}
			} else {
				r.hadError[dep] = true
				msg.Err("Error scanning %s: %s", dep, err)
			}
			continue
		} else {
			if testDeps {
				imps = pkg.TestImports
			} else {
				imps = pkg.Imports
			}

		}

		// Range over all of the identified imports and see which ones we
		// can locate.
		for _, imp := range imps {
			if r.Config.HasIgnore(imp) {
				msg.Debug("Ignoring: %s", imp)
				continue
			}
			pi := r.FindPkg(imp)
			if pi.Loc != LocCgo && pi.Loc != LocGoroot && pi.Loc != LocAppengine {
				msg.Debug("Package %s imports %s", dep, imp)
			}
			switch pi.Loc {
			case LocVendor:
				msg.Debug("In vendor: %s", imp)
				if _, ok := r.alreadyQ[imp]; !ok {
					msg.Debug("Marking %s to be scanned.", imp)
					r.alreadyQ[dep] = true
					queue.PushBack(r.vpath(imp))
					if err := r.Handler.InVendor(imp, addTest); err == nil {
						r.VersionHandler.SetVersion(imp, addTest)
					} else {
						msg.Warn("Error updating %s: %s", imp, err)
					}
				}
			case LocUnknown:
				msg.Debug("Missing %s. Trying to resolve.", imp)
				if ok, err := r.Handler.NotFound(imp, addTest); ok {
					r.alreadyQ[dep] = true
					queue.PushBack(r.vpath(imp))
					r.VersionHandler.SetVersion(imp, addTest)
				} else if err != nil {
					r.hadError[dep] = true
					msg.Warn("Error looking for %s: %s", imp, err)
				} else {
					r.hadError[dep] = true
					msg.Info("Not found: %s (2)", imp)
				}
			case LocGopath:
				msg.Debug("Found on GOPATH, not vendor: %s", imp)
				if _, ok := r.alreadyQ[imp]; !ok {
					// Only scan it if it gets moved into vendor/
					if ok, _ := r.Handler.OnGopath(imp, addTest); ok {
						r.alreadyQ[dep] = true
						queue.PushBack(r.vpath(imp))
						r.VersionHandler.SetVersion(imp, addTest)
					}
				}
			}
		}

	}

	// FIXME: From here to the end is a straight copy of the resolveList() func.
	res := make([]string, 0, queue.Len())

	// In addition to generating a list
	for e := queue.Front(); e != nil; e = e.Next() {
		t := r.Stripv(e.Value.(string))
		root, sp := util.NormalizeName(t)

		// Skip ignored packages
		if r.Config.HasIgnore(e.Value.(string)) {
			msg.Debug("Ignoring: %s", e.Value.(string))
			continue
		}

		// TODO(mattfarina): Need to eventually support devImport
		existing := r.Config.Imports.Get(root)
		if existing == nil && addTest {
			existing = r.Config.DevImports.Get(root)
		}
		if existing != nil {
			if sp != "" && !existing.HasSubpackage(sp) {
				existing.Subpackages = append(existing.Subpackages, sp)
			}
		} else {
			newDep := &cfg.Dependency{
				Name: root,
			}
			if sp != "" {
				newDep.Subpackages = []string{sp}
			}

			if addTest {
				r.Config.DevImports = append(r.Config.DevImports, newDep)
			} else {
				r.Config.Imports = append(r.Config.Imports, newDep)
			}
		}
		res = append(res, t)
	}

	return res, nil
}

// resolveList takes a list and resolves it.
//
// This walks the entire file tree for the given dependencies, not just the
// parts that are imported directly. Using this will discover dependencies
// regardless of OS, and arch.
func (r *Resolver) resolveList(queue *list.List, testDeps, addTest bool) ([]string, error) {
	// When test deps passed in but not resolving return empty.
	if testDeps && !r.ResolveTest {
		return []string{}, nil
	}

	var failedDep string
	for e := queue.Front(); e != nil; e = e.Next() {
		dep := e.Value.(string)
		t := strings.TrimPrefix(dep, r.VendorDir+string(os.PathSeparator))
		if r.Config.HasIgnore(t) {
			msg.Debug("Ignoring: %s", t)
			continue
		}
		r.VersionHandler.Process(t)
		//msg.Warn("#### %s ####", dep)
		//msg.Info("Seen Count: %d", len(r.seen))
		// Catch the outtermost dependency.
		failedDep = dep
		err := filepath.Walk(dep, func(path string, fi os.FileInfo, err error) error {
			if err != nil && err != filepath.SkipDir {
				return err
			}

			// Skip files.
			if !fi.IsDir() {
				return nil
			}
			// Skip dirs that are not source.
			if !srcDir(fi) {
				//msg.Debug("Skip resource %s", fi.Name())
				return filepath.SkipDir
			}

			// Anything that comes through here has already been through
			// the queue.
			r.alreadyQ[path] = true
			e := r.queueUnseen(path, queue, testDeps, addTest)
			if err != nil {
				failedDep = path
				//msg.Err("Failed to fetch dependency %s: %s", path, err)
			}
			return e
		})
		if err != nil && err != filepath.SkipDir {
			msg.Err("Dependency %s failed to resolve: %s.", failedDep, err)
			return []string{}, err
		}
	}

	res := make([]string, 0, queue.Len())

	// In addition to generating a list
	for e := queue.Front(); e != nil; e = e.Next() {
		t := strings.TrimPrefix(e.Value.(string), r.VendorDir+string(os.PathSeparator))
		root, sp := util.NormalizeName(t)

		existing := r.Config.Imports.Get(root)
		if existing == nil && addTest {
			existing = r.Config.DevImports.Get(root)
		}

		if existing != nil {
			if sp != "" && !existing.HasSubpackage(sp) {
				existing.Subpackages = append(existing.Subpackages, sp)
			}
		} else {
			newDep := &cfg.Dependency{
				Name: root,
			}
			if sp != "" {
				newDep.Subpackages = []string{sp}
			}

			if addTest {
				r.Config.DevImports = append(r.Config.DevImports, newDep)
			} else {
				r.Config.Imports = append(r.Config.Imports, newDep)
			}
		}
		res = append(res, e.Value.(string))
	}

	return res, nil
}

// queueUnseenImports scans a package's imports and adds any new ones to the
// processing queue.
func (r *Resolver) queueUnseen(pkg string, queue *list.List, testDeps, addTest bool) error {
	// A pkg is marked "seen" as soon as we have inspected it the first time.
	// Seen means that we have added all of its imports to the list.

	// Already queued indicates that we've either already put it into the queue
	// or intentionally not put it in the queue for fatal reasons (e.g. no
	// buildable source).

	deps, err := r.imports(pkg, testDeps, addTest)
	if err != nil && !strings.HasPrefix(err.Error(), "no buildable Go source") {
		msg.Err("Could not find %s: %s", pkg, err)
		return err
		// NOTE: If we uncomment this, we get lots of "no buildable Go source" errors,
		// which don't ever seem to be helpful. They don't actually indicate an error
		// condition, and it's perfectly okay to run into that condition.
		//} else if err != nil {
		//	msg.Warn(err.Error())
	}

	for _, d := range deps {
		if _, ok := r.alreadyQ[d]; !ok {
			r.alreadyQ[d] = true
			queue.PushBack(d)
		}
	}
	return nil
}

// imports gets all of the imports for a given package.
//
// If the package is in GOROOT, this will return an empty list (but not
// an error).
// If it cannot resolve the pkg, it will return an error.
func (r *Resolver) imports(pkg string, testDeps, addTest bool) ([]string, error) {

	if r.Config.HasIgnore(pkg) {
		msg.Debug("Ignoring %s", pkg)
		return []string{}, nil
	}

	// If this pkg is marked seen, we don't scan it again.
	if _, ok := r.seen[pkg]; ok {
		msg.Debug("Already saw %s", pkg)
		return []string{}, nil
	}

	// FIXME: On error this should try to NotFound to the dependency, and then import
	// it again.
	var imps []string
	p, err := r.BuildContext.ImportDir(pkg, 0)
	if err != nil && strings.HasPrefix(err.Error(), "found packages ") {
		// If we got here it's because a package and multiple packages
		// declared. This is often because of an example with a package
		// or main but +build ignore as a build tag. In that case we
		// try to brute force the packages with a slower scan.
		if testDeps {
			_, imps, err = IterativeScan(pkg)
		} else {
			imps, _, err = IterativeScan(pkg)
		}

		if err != nil {
			return []string{}, err
		}
	} else if err != nil {
		return []string{}, err
	} else {
		if testDeps {
			imps = p.TestImports
		} else {
			imps = p.Imports
		}
	}

	// It is okay to scan a package more than once. In some cases, this is
	// desirable because the package can change between scans (e.g. as a result
	// of a failed scan resolving the situation).
	msg.Debug("=> Scanning %s (%s)", p.ImportPath, pkg)
	r.seen[pkg] = true

	// Optimization: If it's in GOROOT, it has no imports worth scanning.
	if p.Goroot {
		return []string{}, nil
	}

	// We are only looking for dependencies in vendor. No root, cgo, etc.
	buf := []string{}
	for _, imp := range imps {
		if r.Config.HasIgnore(imp) {
			msg.Debug("Ignoring %s", imp)
			continue
		}
		info := r.FindPkg(imp)
		switch info.Loc {
		case LocUnknown:
			// Do we resolve here?
			found, err := r.Handler.NotFound(imp, addTest)
			if err != nil {
				msg.Err("Failed to fetch %s: %s", imp, err)
			}
			if found {
				buf = append(buf, filepath.Join(r.VendorDir, filepath.FromSlash(imp)))
				r.VersionHandler.SetVersion(imp, addTest)
				continue
			}
			r.seen[info.Path] = true
		case LocVendor:
			//msg.Debug("Vendored: %s", imp)
			buf = append(buf, info.Path)
			if err := r.Handler.InVendor(imp, addTest); err == nil {
				r.VersionHandler.SetVersion(imp, addTest)
			} else {
				msg.Warn("Error updating %s: %s", imp, err)
			}
		case LocGopath:
			found, err := r.Handler.OnGopath(imp, addTest)
			if err != nil {
				msg.Err("Failed to fetch %s: %s", imp, err)
			}
			// If the Handler marks this as found, we drop it into the buffer
			// for subsequent processing. Otherwise, we assume that we're
			// in a less-than-perfect, but functional, situation.
			if found {
				buf = append(buf, filepath.Join(r.VendorDir, filepath.FromSlash(imp)))
				r.VersionHandler.SetVersion(imp, addTest)
				continue
			}
			msg.Warn("Package %s is on GOPATH, but not vendored. Ignoring.", imp)
			r.seen[info.Path] = true
		default:
			// Local packages are an odd case. CGO cannot be scanned.
			msg.Debug("===> Skipping %s", imp)
		}
	}

	return buf, nil
}

// sliceToQueue is a special-purpose function for unwrapping a slice of
// dependencies into a queue of fully qualified paths.
func sliceToQueue(deps []*cfg.Dependency, basepath string) *list.List {
	l := list.New()
	for _, e := range deps {
		if len(e.Subpackages) > 0 {
			for _, v := range e.Subpackages {
				ip := e.Name
				if v != "." && v != "" {
					ip = ip + "/" + v
				}
				msg.Debug("Adding local Import %s to queue", ip)
				l.PushBack(filepath.Join(basepath, filepath.FromSlash(ip)))
			}
		} else {
			msg.Debug("Adding local Import %s to queue", e.Name)
			l.PushBack(filepath.Join(basepath, filepath.FromSlash(e.Name)))
		}

	}
	return l
}

// PkgLoc describes the location of the package.
type PkgLoc uint8

const (
	// LocUnknown indicates the package location is unknown (probably not present)
	LocUnknown PkgLoc = iota
	// LocLocal inidcates that the package is in a local dir, not GOPATH or GOROOT.
	LocLocal
	// LocVendor indicates that the package is in a vendor/ dir
	LocVendor
	// LocGopath inidcates that the package is in GOPATH
	LocGopath
	// LocGoroot indicates that the package is in GOROOT
	LocGoroot
	// LocCgo indicates that the package is a a CGO package
	LocCgo
	// LocAppengine indicates the package is part of the appengine SDK. It's a
	// special build mode. https://blog.golang.org/the-app-engine-sdk-and-workspaces-gopath
	// Why does a Google product get a special case build mode with a local
	// package?
	LocAppengine
	// LocRelative indicates the package is a relative directory
	LocRelative
)

// PkgInfo represents metadata about a package found by the resolver.
type PkgInfo struct {
	Name, Path string
	Vendored   bool
	Loc        PkgLoc
}

// FindPkg takes a package name and attempts to find it on the filesystem
//
// The resulting PkgInfo will indicate where it was found.
func (r *Resolver) FindPkg(name string) *PkgInfo {
	// We cachae results for FindPkg to reduce the number of filesystem ops
	// that we have to do. This is a little risky because certain directories,
	// like GOPATH, can be modified while we're running an operation, and
	// render the cache inaccurate.
	//
	// Unfound items (LocUnknown) are never cached because we assume that as
	// part of the response, the Resolver may fetch that dependency.
	if i, ok := r.findCache[name]; ok {
		//msg.Info("Cache hit on %s", name)
		return i
	}

	// 502 individual packages scanned.
	// No cache:
	// glide -y etcd.yaml list  0.27s user 0.19s system 85% cpu 0.534 total
	// With cache:
	// glide -y etcd.yaml list  0.22s user 0.15s system 85% cpu 0.438 total

	var p string
	info := &PkgInfo{
		Name: name,
	}

	if strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") {
		info.Loc = LocRelative
		r.findCache[name] = info
		return info
	}

	// Check _only_ if this dep is in the current vendor directory.
	p = filepath.Join(r.VendorDir, filepath.FromSlash(name))
	if pkgExists(p) {
		info.Path = p
		info.Loc = LocVendor
		info.Vendored = true
		r.findCache[name] = info
		return info
	}

	// TODO: Do we need this if we always flatten?
	// Recurse backward to scan other vendor/ directories
	//for wd := cwd; wd != "/"; wd = filepath.Dir(wd) {
	//p = filepath.Join(wd, "vendor", filepath.FromSlash(name))
	//if fi, err = os.Stat(p); err == nil && (fi.IsDir() || isLink(fi)) {
	//info.Path = p
	//info.PType = ptypeVendor
	//info.Vendored = true
	//return info
	//}
	//}

	// Check $GOPATH
	for _, rr := range filepath.SplitList(r.BuildContext.GOPATH) {
		p = filepath.Join(rr, "src", filepath.FromSlash(name))
		if pkgExists(p) {
			info.Path = p
			info.Loc = LocGopath
			r.findCache[name] = info
			return info
		}
	}

	// Check $GOROOT
	for _, rr := range filepath.SplitList(r.BuildContext.GOROOT) {
		p = filepath.Join(rr, "src", filepath.FromSlash(name))
		if pkgExists(p) {
			info.Path = p
			info.Loc = LocGoroot
			r.findCache[name] = info
			return info
		}
	}

	// If this is "C", we're dealing with cgo
	if name == "C" {
		info.Loc = LocCgo
		r.findCache[name] = info
	} else if name == "appengine" || name == "appengine_internal" ||
		strings.HasPrefix(name, "appengine/") ||
		strings.HasPrefix(name, "appengine_internal/") {
		// Appengine is a special case when it comes to Go builds. It is a local
		// looking package only available within appengine. It's a special case
		// where Google products are playing with each other.
		// https://blog.golang.org/the-app-engine-sdk-and-workspaces-gopath
		info.Loc = LocAppengine
		r.findCache[name] = info
	} else if name == "context" || name == "net/http/httptrace" {
		// context and net/http/httptrace are packages being added to
		// the Go 1.7 standard library. Some packages, such as golang.org/x/net
		// are importing it with build flags in files for go1.7. Need to detect
		// this and handle it.
		info.Loc = LocGoroot
		r.findCache[name] = info
	}

	return info
}

func pkgExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && (fi.IsDir() || isLink(fi))
}

// isLink returns true if the given FileInfo is a symbolic link.
func isLink(fi os.FileInfo) bool {
	return fi.Mode()&os.ModeSymlink == os.ModeSymlink
}

// IsSrcDir returns true if this is a directory that could have source code,
// false otherwise.
//
// Directories with _ or . prefixes are skipped, as are testdata and vendor.
func IsSrcDir(fi os.FileInfo) bool {
	return srcDir(fi)
}

func srcDir(fi os.FileInfo) bool {
	if !fi.IsDir() {
		return false
	}

	// Ignore _foo and .foo
	if strings.HasPrefix(fi.Name(), "_") || strings.HasPrefix(fi.Name(), ".") {
		return false
	}

	// Ignore testdata. For now, ignore vendor.
	if fi.Name() == "testdata" || fi.Name() == "vendor" {
		return false
	}

	return true
}

// checkForBasedirSymlink checks to see if the given basedir is actually a
// symlink. In the case that it is a symlink, the symlink is read and returned.
// If the basedir is not a symlink, the provided basedir argument is simply
// returned back to the caller.
func checkForBasedirSymlink(basedir string) (string, error) {
	fi, err := os.Lstat(basedir)
	if err != nil {
		return "", err
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		return os.Readlink(basedir)
	}

	return basedir, nil
}
