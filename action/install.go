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
	"github.com/sdboyer/vsolver"
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
	sm, err := vsolver.NewSourceManager(filepath.Join(installer.Home, "cache"), base, false, dependency.Analyzer{})
	defer sm.Release()
	if err != nil {
		msg.Err(err.Error())
		return
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

	var s vsolver.Solver
	if gpath.HasLock(base) {
		args.L, err = LoadLockfile(base, conf)
		if err != nil {
			msg.Err("Could not load lockfile.")
			return
		}

		s, err = vsolver.Prepare(args, opts, sm)
		if err != nil {
			msg.Err("Could not set up solver: %s", err)
			return
		}
		digest, err := s.HashInputs()

		// Check if digests match, and warn if they don't
		if bytes.Equal(digest, args.L.InputHash()) {
			if so {
				msg.Err("glide.yaml is out of sync with glide.lock")
				return
			} else {
				msg.Warn("glide.yaml is out of sync with glide.lock!")
			}
		}
		err = writeVendor(vend, args.L, sm, sv)
		if err != nil {
			msg.Err(err.Error())
			return
		}
	} else if io || so {
		msg.Err("No glide.lock file could be found.")
		return
	} else {
		// There is no lock, so we have to solve first
		s, err = vsolver.Prepare(args, opts, sm)
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
	}
}

// TODO This will almost certainly need to be renamed and move somewhere else
func writeVendor(vendor string, l vsolver.Lock, sm vsolver.SourceManager, strip bool) error {
	td, err := ioutil.TempDir(os.TempDir(), "glide")
	if err != nil {
		return fmt.Errorf("Error while creating temp dir for vendor directory: %s", err)
	}
	defer os.RemoveAll(td)

	err = vsolver.CreateVendorTree(td, l, sm, strip)
	if err != nil {
		return fmt.Errorf("Error while generating vendor tree: %s", err)
	}

	// Move the existing vendor dir to somewhere safe while we put the new one
	// in order to provide insurance against errors for as long as possible
	td2, err := ioutil.TempDir(filepath.Dir(vendor), "vendor")
	if err != nil {
		return fmt.Errorf("Error creating swap dir for existing vendor directory: %s", err)
	}

	err = os.Rename(vendor, td2)
	defer os.RemoveAll(td2)
	if err != nil {
		return fmt.Errorf("Error moving existing vendor into swap dir: %s", err)
	}

	err = os.Rename(td, vendor)
	if err != nil {
		return fmt.Errorf("Error while moving generated vendor directory into place: %s", err)
	}

	return nil
}

// LoadLockfile loads the contents of a glide.lock file.
//
// TODO: This should go in another package.
func LoadLockfile(base string, conf *cfg.Config) (*cfg.Lockfile, error) {
	yml, err := ioutil.ReadFile(filepath.Join(base, gpath.LockFile))
	if err != nil {
		return nil, err
	}
	lock, err := cfg.LockfileFromYaml(yml)
	if err != nil {
		return nil, err
	}

	hash, err := conf.Hash()
	if err != nil {
		return nil, err
	}

	if hash != lock.Hash {
		msg.Warn("Lock file may be out of date. Hash check of YAML failed. You may need to run 'update'")
	}

	return lock, nil
}
