package action

import (
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/dependency"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
	"github.com/sdboyer/vsolver"
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

	args := vsolver.SolveArgs{
		N:    vsolver.ProjectName(conf.ProjectName),
		Root: filepath.Dir(vend),
		M:    conf,
	}

	opts := vsolver.SolveOpts{
		Trace:       true,
		TraceLogger: log.New(os.Stdout, "", 0),
	}

	if len(projs) == 0 {
		opts.ChangeAll = true
	} else {
		opts.ChangeAll = false
		for _, p := range projs {
			if !conf.HasDependency(p) {
				msg.Die("Cannot update %s, as it is not listed as dependency in glide.yaml.", p)
			}
			opts.ToChange = append(opts.ToChange, vsolver.ProjectName(p))
		}
	}

	if gpath.HasLock(base) {
		args.L, err = LoadLockfile(base, conf)
		if err != nil {
			msg.Err("Could not load lockfile, aborting: %s", err)
			return
		}
	}

	// Create the SourceManager for this run
	sm, err := vsolver.NewSourceManager(filepath.Join(installer.Home, "cache"), base, false, dependency.Analyzer{})
	defer sm.Release()
	if err != nil {
		msg.Err(err.Error())
		return
	}

	// Prepare a solver. This validates our args and opts.
	s, err := vsolver.Prepare(args, opts, sm)
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

	err = writeVendor(vend, r, sm, sv)
	if err != nil {
		msg.Err(err.Error())
		return
	}

	// Create and write out a new lock file from the result
	lf := &cfg.Lockfile{
		Hash:    hex.EncodeToString(r.InputHash()),
		Updated: time.Now(),
	}

	for _, p := range r.Projects() {
		pi := p.Ident()
		l := &cfg.Lock{
			Name:    string(pi.LocalName),
			VcsType: "", // TODO allow this to be extracted from sm
		}

		if l.Name != pi.NetworkName && pi.NetworkName != "" {
			l.Repository = pi.NetworkName
		}

		v := p.Version()
		if pv, ok := v.(vsolver.PairedVersion); ok {
			l.Version = pv.Underlying().String()
		} else {
			l.Version = v.String()
		}

		lf.Imports = append(lf.Imports, l)
	}

	wl := true
	if args.L != nil {
		f1, err := args.L.(*cfg.Lockfile).Fingerprint()
		f2, err2 := lf.Fingerprint()
		if err == nil && err2 == nil && f1 == f2 {
			wl = false
		}
	}

	if wl {
		if err := lf.WriteFile(filepath.Join(base, gpath.LockFile)); err != nil {
			msg.Err("Could not write lock file to %s: %s", base, err)
			return
		}
	} else {
		msg.Info("Versions did not change. Skipping glide.lock update.")
	}

	err = writeVendor(vend, r, sm, sv)
	if err != nil {
		msg.Err(err.Error())
	}
}
