package action

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

// Install installs a vendor directory based on an existing Glide configuration.
func Install(installer *repo.Installer, io, so, sv bool) {
	base := "."
	// Ensure GOPATH
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()

	// TODO might need a better way for discovering the root
	vend, err := gpath.Vendor()
	if err != nil {
		msg.Die("Could not find the vendor dir: %s", err)
	}

	// Create the SourceManager for this run
	sm, err := gps.NewSourceManager(dependency.Analyzer{}, filepath.Join(installer.Home, "cache"), false)
	defer sm.Release()
	if err != nil {
		msg.Err(err.Error())
		return
	}

	params := gps.SolveParameters{
		RootDir:     filepath.Dir(vend),
		ImportRoot:  gps.ProjectRoot(conf.ProjectRoot),
		Manifest:    conf,
		Trace:       true,
		TraceLogger: log.New(os.Stdout, "", 0),
	}

	var s gps.Solver
	if gpath.HasLock(base) {
		params.Lock, err = loadLockfile(base, conf)
		if err != nil {
			msg.Err("Could not load lockfile.")
			return
		}

		s, err = gps.Prepare(params, sm)
		if err != nil {
			msg.Err("Could not set up solver: %s", err)
			return
		}
		digest, err := s.HashInputs()

		// Check if digests match, and warn if they don't
		if bytes.Equal(digest, params.Lock.InputHash()) {
			if so {
				msg.Err("glide.yaml is out of sync with glide.lock")
				return
			} else {
				msg.Warn("glide.yaml is out of sync with glide.lock!")
			}
		}

		gw := safeGroupWriter{
			resultLock:  params.Lock,
			vendor:      vend,
			sm:          sm,
			stripVendor: sv,
		}

		err = gw.writeAllSafe()
		if err != nil {
			msg.Err(err.Error())
			return
		}
	} else if io || so {
		msg.Err("No glide.lock file could be found.")
		return
	} else {
		// There is no lock, so we have to solve first
		s, err = gps.Prepare(params, sm)
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
			resultLock:  r,
			vendor:      vend,
			sm:          sm,
			stripVendor: sv,
		}

		err = gw.writeAllSafe()
		if err != nil {
			msg.Err(err.Error())
			return
		}
	}
}

// locksAreEquivalent compares the fingerprints between two locks to determine
// if they're equivalent.
//
// If the either of the locks are nil, the input hashes are different, the
// fingerprints are different, or any error is returned from fingerprinting,
// this function returns false.
func locksAreEquivalent(l1, l2 *cfg.Lockfile) bool {
	if l1 != nil && l2 != nil {
		if l1.Hash != l2.Hash {
			return false
		}

		f1, err := l1.Fingerprint()
		f2, err2 := l2.Fingerprint()
		if err == nil && err2 == nil && f1 == f2 {
			return true
		}
	}
	return false
}

// safeGroupWriter provides a slipshod-but-better-than-nothing approach to
// grouping together yaml, lock, and vendor dir writes.
type safeGroupWriter struct {
	conf              *cfg.Config
	lock              *cfg.Lockfile
	resultLock        gps.Lock
	sm                gps.SourceManager
	glidefile, vendor string
	stripVendor       bool
}

// writeAllSafe writes out some combination of config yaml, lock, and a vendor
// tree, to a temp dir, then moves them into place if and only if all the write
// operations succeeded. It also does its best to roll back if any moves fail.
//
// This helps to ensure glide doesn't exit with a partial write, resulting in an
// undefined disk state.
//
// - If a gw.conf is provided, it will be written to gw.glidefile
// - If gw.lock is provided without a gw.resultLock, it will be written to
//   `glide.lock` in the parent dir of gw.vendor
// - If gw.lock and gw.resultLock are both provided and are not equivalent,
//   the resultLock will be written to the same location as above, and a vendor
//   tree will be written to gw.vendor
// - If gw.resultLock is provided and gw.lock is not, it will write both a lock
//   and vendor dir in the same way
//
// Any of the conf, lock, or result can be omitted; the grouped write operation
// will continue for whichever inputs are present.
func (gw safeGroupWriter) writeAllSafe() error {
	// Decide which writes we need to do
	var writeConf, writeLock, writeVendor bool

	if gw.conf != nil {
		writeConf = true
	}

	if gw.resultLock != nil {
		if gw.lock == nil {
			writeLock, writeVendor = true, true
		} else {
			rlf := cfg.LockfileFromSolverLock(gw.resultLock)
			if !locksAreEquivalent(rlf, gw.lock) {
				writeLock, writeVendor = true, true
			}
		}
	} else if gw.lock != nil {
		writeLock = true
	}

	if !writeConf && !writeLock && !writeVendor {
		// nothing to do
		return nil
	}

	if writeConf && gw.glidefile == "" {
		return fmt.Errorf("Must provide a path if writing out a config yaml.")
	}

	if (writeLock || writeVendor) && gw.vendor == "" {
		return fmt.Errorf("Must provide a vendor dir if writing out a lock or vendor dir.")
	}

	if writeVendor && gw.sm == nil {
		return fmt.Errorf("Must provide a SourceManager if writing out a vendor dir.")
	}

	td, err := ioutil.TempDir(os.TempDir(), "glide")
	if err != nil {
		return fmt.Errorf("Error while creating temp dir for vendor directory: %s", err)
	}
	defer os.RemoveAll(td)

	if writeConf {
		if err := gw.conf.WriteFile(filepath.Join(td, "glide.yaml")); err != nil {
			return fmt.Errorf("Failed to write glide YAML file: %s", err)
		}
	}

	if writeLock {
		if gw.resultLock == nil {
			// the result lock is nil but the flag is on, so we must be writing
			// the other one
			if err := gw.lock.WriteFile(filepath.Join(td, gpath.LockFile)); err != nil {
				return fmt.Errorf("Failed to write glide lock file: %s", err)
			}
		} else {
			rlf := cfg.LockfileFromSolverLock(gw.resultLock)
			if err := rlf.WriteFile(filepath.Join(td, gpath.LockFile)); err != nil {
				return fmt.Errorf("Failed to write glide lock file: %s", err)
			}
		}
	}

	if writeVendor {
		err = gps.WriteDepTree(filepath.Join(td, "vendor"), gw.resultLock, gw.sm, gw.stripVendor)
		if err != nil {
			return fmt.Errorf("Error while generating vendor tree: %s", err)
		}
	}

	// Move the existing files and dirs to the temp dir while we put the new
	// ones in, to provide insurance against errors for as long as possible
	var fail bool
	var failerr error
	type pathpair struct {
		from, to string
	}
	var restore []pathpair

	if writeConf {
		if _, err := os.Stat(gw.glidefile); err == nil {
			// move out the old one
			tmploc := filepath.Join(td, "glide.yaml-old")
			failerr = os.Rename(gw.glidefile, tmploc)
			if failerr != nil {
				fail = true
			} else {
				restore = append(restore, pathpair{from: tmploc, to: gw.glidefile})
			}
		}

		// move in the new one
		failerr = os.Rename(filepath.Join(td, "glide.yaml"), gw.glidefile)
		if failerr != nil {
			fail = true
		}
	}

	if !fail && writeLock {
		tgt := filepath.Join(filepath.Dir(gw.vendor), gpath.LockFile)
		if _, err := os.Stat(tgt); err == nil {
			// move out the old one
			tmploc := filepath.Join(td, "glide.lock-old")

			failerr = os.Rename(tgt, tmploc)
			if failerr != nil {
				fail = true
			} else {
				restore = append(restore, pathpair{from: tmploc, to: tgt})
			}
		}

		// move in the new one
		failerr = os.Rename(filepath.Join(td, gpath.LockFile), tgt)
		if failerr != nil {
			fail = true
		}
	}

	// have to declare out here so it's present later
	var vendorbak string
	if !fail && writeVendor {
		if _, err := os.Stat(gw.vendor); err == nil {
			// move out the old vendor dir. just do it into an adjacent dir, in
			// order to mitigate the possibility of a pointless cross-filesystem move
			vendorbak = gw.vendor + "-old"
			if _, err := os.Stat(vendorbak); err == nil {
				// Just in case that happens to exist...
				vendorbak = filepath.Join(td, "vendor-old")
			}
			failerr = os.Rename(gw.vendor, vendorbak)
			if failerr != nil {
				fail = true
			} else {
				restore = append(restore, pathpair{from: vendorbak, to: gw.vendor})
			}
		}

		// move in the new one
		failerr = os.Rename(filepath.Join(td, "vendor"), gw.vendor)
		if failerr != nil {
			fail = true
		}
	}

	// If we failed at any point, move all the things back into place, then bail
	if fail {
		for _, pair := range restore {
			// Nothing we can do on err here, we're already in recovery mode
			os.Rename(pair.from, pair.to)
		}
		return failerr
	}

	// Renames all went smoothly. The deferred os.RemoveAll will get the temp
	// dir, but if we wrote vendor, we have to clean that up directly

	if writeVendor {
		// Again, kinda nothing we can do about an error at this point
		os.RemoveAll(vendorbak)
	}

	return nil
}

// loadLockfile loads the contents of a glide.lock file.
func loadLockfile(base string, conf *cfg.Config) (*cfg.Lockfile, error) {
	yml, err := ioutil.ReadFile(filepath.Join(base, gpath.LockFile))
	if err != nil {
		return nil, err
	}
	lock, err := cfg.LockfileFromYaml(yml)
	if err != nil {
		return nil, err
	}

	return lock, nil
}
