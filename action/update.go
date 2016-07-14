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

// Update updates repos and the lock file from the main glide yaml.
func Update(installer *repo.Installer, sv bool, projs []string) {
	base := "."
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()

	// TODO(mattfarina): Detect when a new dependency has been added or removed
	// from the project. A removed dependency should warn and an added dependency
	// should be added to the glide.yaml file. See issue #193.

	// TODO might need a better way for discovering the root
	vend, err := gpath.Vendor()
	if err != nil {
		msg.Die("Could not find the vendor dir: %s", err)
	}

	params := gps.SolveParameters{
		RootDir:     filepath.Dir(vend),
		ImportRoot:  gps.ProjectRoot(conf.ProjectRoot),
		Manifest:    conf,
		Ignore:      conf.Ignore,
		Trace:       true,
		TraceLogger: log.New(os.Stdout, "", 0),
	}

	if len(projs) == 0 {
		params.ChangeAll = true
	} else {
		params.ChangeAll = false
		for _, p := range projs {
			if !conf.HasDependency(p) {
				msg.Die("Cannot update %s, as it is not listed as dependency in glide.yaml.", p)
			}
			params.ToChange = append(params.ToChange, gps.ProjectRoot(p))
		}
	}

	if gpath.HasLock(base) {
		params.Lock, err = loadLockfile(base, conf)
		if err != nil {
			msg.Err("Could not load lockfile, aborting: %s", err)
			return
		}
	}

	// Create the SourceManager for this run
	sm, err := gps.NewSourceManager(dependency.Analyzer{}, filepath.Join(installer.Home, "cache"), false)
	if err != nil {
		msg.Err(err.Error())
		return
	}
	defer sm.Release()

	// Prepare a solver. This validates our params.
	s, err := gps.Prepare(params, sm)
	if err != nil {
		msg.Err("Could not set up solver: %s", err)
		return
	}

	r, err := s.Solve()
	if err != nil {
		// TODO better error handling
		msg.Err(err.Error())
		return
	}

	gw := safeGroupWriter{
		lock:        params.Lock.(*cfg.Lockfile),
		resultLock:  r,
		sm:          sm,
		vendor:      vend,
		stripVendor: sv,
	}

	err = gw.writeAllSafe()
	if err != nil {
		msg.Err(err.Error())
		return
	}
}
