// Package overrides handles managing overrides in the running application
package overrides

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

var overrides map[string]*override

func init() {
	overrides = make(map[string]*override)
}

type override struct {
	Repo, Vcs string
}

// Get retrieves informtion about an override. It returns.
// - bool if found
// - new repo location
// - vcs type
func Get(k string) (bool, string, string) {
	o, f := overrides[k]
	if !f {
		return false, "", ""
	}

	return true, o.Repo, o.Vcs
}

// Load pulls the overrides into memory
func Load() error {
	home := gpath.Home()

	op := filepath.Join(home, "overrides.yaml")

	var ov *Overrides
	if _, err := os.Stat(op); os.IsNotExist(err) {
		msg.Debug("No overrides.yaml file exists")
		ov = &Overrides{
			Repos: make(OverrideRepos, 0),
		}
	} else {
		ov, err = ReadOverridesFile(op)
		if err != nil {
			return fmt.Errorf("Error reading existing overrides.yaml file: %s", err)
		}
	}

	msg.Info("Loading overrides from overrides.yaml file")
	for _, o := range ov.Repos {
		msg.Debug("Found override: %s to %s (%s)", o.Original, o.Repo, o.Vcs)
		no := &override{
			Repo: o.Repo,
			Vcs:  o.Vcs,
		}
		overrides[o.Original] = no
	}

	return nil
}
