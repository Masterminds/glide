package repo

import (
	"sync"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
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

	for _, dep := range newConf.Imports {
		wg.Add(1)
		in <- dep
	}

	wg.Wait()
	// Close goroutines setting the version
	for ii := 0; ii < concurrentWorkers; ii++ {
		done <- struct{}{}
	}

	return newConf, nil
}
