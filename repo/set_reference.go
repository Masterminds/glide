package repo

import (
	"sync"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	"github.com/codegangsta/cli"
)

// SetReference is a command to set the VCS reference (commit id, tag, etc) for
// a project.
func SetReference(conf *cfg.Config, resolveTest bool) error {

	if len(conf.Imports) == 0 && len(conf.DevImports) == 0 {
		msg.Info("No references set.\n")
		return nil
	}

	done := make(chan struct{}, concurrentWorkers)
	in := make(chan *cfg.Dependency, concurrentWorkers)
	var wg sync.WaitGroup
	var lock sync.Mutex
	var returnErr error

	for i := 0; i < concurrentWorkers; i++ {
		go func(ch <-chan *cfg.Dependency) {
			for {
				select {
				case dep := <-ch:

					var loc string
					if dep.Repository != "" {
						loc = dep.Repository
					} else {
						loc = "https://" + dep.Name
					}
					key, err := cache.Key(loc)
					if err != nil {
						msg.Die(err.Error())
					}
					cache.Lock(key)
					if err := VcsVersion(dep); err != nil {
						msg.Err("Failed to set version on %s to %s: %s\n", dep.Name, dep.Reference, err)

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
			wg.Add(1)
			in <- dep
		}
	}

	if resolveTest {
		for _, dep := range conf.DevImports {
			if !conf.HasIgnore(dep.Name) {
				wg.Add(1)
				in <- dep
			}
		}
	}

	wg.Wait()
	// Close goroutines setting the version
	for i := 0; i < concurrentWorkers; i++ {
		done <- struct{}{}
	}
	// close(done)
	// close(in)

	return returnErr
}
