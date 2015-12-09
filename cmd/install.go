package cmd

import (
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/glide/cfg"
)

// LockFileExists checks if a lock file exists. If not it jumps to the update
// command.
func LockFileExists(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fname := p.Get("filename", "glide.lock").(string)
	if _, err := os.Stat(fname); err != nil {
		Info("Lock file (glide.lock) does not exist. Performing update.")
		return false, &cookoo.Reroute{"update"}
	}

	return true, nil
}

// LoadLockFile loads the lock file to the context and checks if it is correct
// for the loaded cfg file.
func LoadLockFile(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fname := p.Get("filename", "glide.lock").(string)
	conf := p.Get("conf", nil).(*cfg.Config)

	yml, err := ioutil.ReadFile(fname)
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
		return nil, errors.New("Lock file does not match YAML configuration. Consider running 'update'")
	}

	return lock, nil
}

// Install installs the dependencies from a Lockfile.
func Install(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	lock := p.Get("lock", nil).(*cfg.Lockfile)
	conf := p.Get("conf", nil).(*cfg.Config)
	force := p.Get("force", true).(bool)
	home := p.Get("home", "").(string)
	cache := p.Get("cache", false).(bool)
	cacheGopath := p.Get("cacheGopath", false).(bool)
	useGopath := p.Get("useGopath", false).(bool)

	cwd, err := VendorPath(c)
	if err != nil {
		return false, err
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
		Info("No dependencies found. Nothing installed.\n")
		return false, nil
	}

	// for _, dep := range newConf.Imports {
	// 	if err := VcsUpdate(dep, cwd, home, force, cache, cacheGopath, useGopath); err != nil {
	// 		Warn("Update failed for %s: %s\n", dep.Name, err)
	// 	}
	// }

	done := make(chan struct{}, concurrentWorkers)
	in := make(chan *cfg.Dependency, concurrentWorkers)
	var wg sync.WaitGroup

	for i := 0; i < concurrentWorkers; i++ {
		go func(ch <-chan *cfg.Dependency) {
			for {
				select {
				case dep := <-ch:
					if err := VcsUpdate(dep, cwd, home, force, cache, cacheGopath, useGopath); err != nil {
						Warn("Update failed for %s: %s\n", dep.Name, err)
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
	for i := 0; i < concurrentWorkers; i++ {
		done <- struct{}{}
	}

	return newConf, nil
}
