package cfg

import (
	"time"
)

// Lockfile represents a glide.lock file.
type Lockfile struct {
	Hash       string    `yaml:"hash"`
	Updated    time.Time `yaml:"updated"`
	Imports    []*Lock   `yaml:"imports"`
	DevImports []*Lock   `yaml:"devImports"`
}

type Lock struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func NewLockfile(ds []*Dependency) *Lockfile {
	lf := &Lockfile{
		Updated: time.Now(),
		Imports: make([]*Lock, len(ds)),
	}

	for i := 0; i < len(ds); i++ {
		lf.Imports[i] = &Lock{
			Name:    ds[i].Name,
			Version: ds[i].Reference,
		}
	}

	return lf
}

func LockfileFromMap(ds map[string]*Dependency) *Lockfile {
	lf := &Lockfile{
		Updated: time.Now(),
		Imports: make([]*Lock, len(ds)),
	}

	i := 0
	for name, dep := range ds {
		lf.Imports[i] = &Lock{
			Name:    name,
			Version: dep.Pin,
		}
		i++
	}

	return lf
}
