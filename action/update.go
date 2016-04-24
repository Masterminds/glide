package action

import (
	"encoding/hex"
	"path/filepath"
	"time"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
	"github.com/Sirupsen/logrus"
	"github.com/sdboyer/vsolver"
)

// Update updates repos and the lock file from the main glide yaml.
func Update(installer *repo.Installer, skipRecursive, strip, stripVendor bool) {
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

	// Create the SourceManager for this run
	sm, err := vsolver.NewSourceManager(filepath.Join(installer.Home, "cache"), base, true, false, dependency.Analyzer{})
	if err != nil {
		msg.Die(err.Error())
	}
	// TODO this defer doesn't trigger when we exit through a msg.Die() call
	defer sm.Release()

	opts := vsolver.SolveOpts{
		N:    vsolver.ProjectName(conf.ProjectName),
		Root: filepath.Dir(vend),
		M:    conf,
	}

	if gpath.HasLock(base) {
		opts.L, err = LoadLockfile(base, conf)
		if err != nil {
			msg.Warn("Could not load lockfile; all projects will be updated.")
		}
	}

	s := vsolver.NewSolver(sm, logrus.New())
	r, err := s.Solve(opts)
	if err != nil {
		// TODO better error handling
		sm.Release()
		msg.Die(err.Error())
	}

	err = writeVendor(vend, r, sm)
	if err != nil {
		sm.Release()
		msg.Die(err.Error())
	}

	// TODO compare old and new lock, and only change if contents differ

	// Create and write out a new lock file from the result
	lf := &cfg.Lockfile{
		Hash:    hex.EncodeToString(r.InputHash()),
		Updated: time.Now(),
	}

	for _, p := range r.Projects() {
		l := &cfg.Lock{
			Name:       string(p.Name()),
			Repository: p.URI(), // TODO this is wrong
			VcsType:    "",      // TODO allow this to be extracted from sm
		}

		v := p.Version()
		if pv, ok := v.(vsolver.PairedVersion); ok {
			l.Version = pv.Underlying().String()
		} else {
			l.Version = pv.String()
		}
	}

	err = lf.WriteFile(filepath.Join(base, gpath.LockFile))
	if err != nil {
		sm.Release()
		msg.Die("Error on writing new lock file: %s", err)
	}

	err = writeVendor(vend, r, sm)
	if err != nil {
		sm.Release()
		msg.Die(err.Error())
	}
}
