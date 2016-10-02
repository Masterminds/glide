package action

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
	"github.com/sdboyer/gps"
)

// Remove removes a constraint from the configuration.
func Remove(packages []string, inst *repo.Installer, stripVendor bool) {
	base := gpath.Basepath()
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()

	glidefile, err := gpath.Glide()
	if err != nil {
		msg.Die("Could not find Glide file: %s", err)
	}

	vend, err := gpath.Vendor()
	if err != nil {
		msg.Die("Could not find the vendor dir: %s", err)
	}

	rd := filepath.Dir(glidefile)
	rt, err := gps.ListPackages(rd, conf.Name)
	if err != nil {
		msg.Die("Error while scanning project imports: %s", err)
	}

	params := gps.SolveParameters{
		RootDir:         rd,
		RootPackageTree: rt,
		Manifest:        conf,
		Trace:           true,
		TraceLogger:     log.New(os.Stdout, "", 0),
	}

	// We load the lock file early and bail out if there's a problem, because we
	// don't want a get to just update all deps without the user explictly
	// making that choice.
	if gpath.HasLock(base) {
		params.Lock, _, err = loadLockfile(base, conf)
		if err != nil {
			msg.Err("Could not load lockfile; aborting get. Other dependency versions cannot be preserved without an invalid lock file . Error was: %s", err)
			return
		}
	}

	// Create the SourceManager for this run
	sm, err := gps.NewSourceManager(dependency.Analyzer{}, filepath.Join(inst.Home, "cache"))
	defer sm.Release()
	if err != nil {
		msg.Err(err.Error())
		return
	}

	msg.Info("Preparing to remove %d packages.", len(packages))
	conf.Imports = rmDeps(packages, conf.Imports)
	conf.DevImports = rmDeps(packages, conf.DevImports)

	// Prepare a solver. This validates our params.
	s, err := gps.Prepare(params, sm)
	if err != nil {
		msg.Err("Aborted get - could not set up solver to reconcile dependencies: %s", err)
		return
	}

	r, err := s.Solve()
	if err != nil {
		msg.Err("Failed to find a solution for all new dependencies: %s", err.Error())
		return
	}

	// Solve succeeded. Write out the yaml, lock, and vendor to a tmpdir, then mv
	// them all into place iff all the writes worked

	gw := safeGroupWriter{
		conf:        conf,
		lock:        params.Lock.(*cfg.Lockfile),
		resultLock:  r,
		sm:          sm,
		glidefile:   glidefile,
		vendor:      vend,
		stripVendor: stripVendor,
	}

	err = gw.writeAllSafe()
	if err != nil {
		msg.Err(err.Error())
	}
}

// rmDeps returns a list of dependencies that do not contain the given pkgs.
//
// It generates neither an error nor a warning for a pkg that does not exist
// in the list of deps.
func rmDeps(pkgs []string, deps []*cfg.Dependency) []*cfg.Dependency {
	res := []*cfg.Dependency{}
	for _, d := range deps {
		rem := false
		for _, p := range pkgs {
			if p == d.Name {
				rem = true
			}
		}
		if !rem {
			res = append(res, d)
		}
	}
	return res
}
