package repo

import (
	"sync"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

// SetReference is a command to set the VCS reference (commit id, tag, etc) for
// a project.
func SetReference(conf *cfg.Config, resolveTest bool) error {

	cwd, err := gpath.Vendor()
	if err != nil {
		return err
	}

	if len(conf.Imports) == 0 && len(conf.DevImports) == 0 {
		msg.Info("No references set.\n")
		return nil
	}

	done := make(chan struct{}, concurrentWorkers)
	in := make(chan *cfg.Dependency, concurrentWorkers)
	var wg sync.WaitGroup

	for i := 0; i < concurrentWorkers; i++ {
		go func(ch <-chan *cfg.Dependency) {
			for {
				select {
				case dep := <-ch:
					if err := VcsVersion(dep, cwd); err != nil {
						msg.Err("Failed to set version on %s to %s: %s\n", dep.Name, dep.Reference, err)
					}
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

	return nil
}
