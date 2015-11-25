package cfg

import (
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Lockfile represents a glide.lock file.
type Lockfile struct {
	Hash       string    `yaml:"hash"`
	Updated    time.Time `yaml:"updated"`
	Imports    Locks     `yaml:"imports"`
	DevImports Locks     `yaml:"devImports"`
}

// Marshal converts a Config instance to YAML
func (lf *Lockfile) Marshal() ([]byte, error) {
	yml, err := yaml.Marshal(&lf)
	if err != nil {
		return []byte{}, err
	}
	return yml, nil
}

type Locks []*Lock

// Len returns the length of the Locks. This is needed for sorting with
// the sort package.
func (l Locks) Len() int {
	return len(l)
}

// Less is needed for the sort interface. It compares two locks based on
// their name.
func (l Locks) Less(i, j int) bool {

	// Names are normalized to lowercase because case affects sorting order. For
	// example, Masterminds comes before kylelemons. Making them lowercase
	// causes kylelemons to come first which is what is expected.
	return strings.ToLower(l[i].Name) < strings.ToLower(l[j].Name)
}

// Swap is needed for the sort interface. It swaps the position of two
// locks.
func (l Locks) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type Lock struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func NewLockfile(ds Dependencies, hash string) *Lockfile {
	lf := &Lockfile{
		Hash:    hash,
		Updated: time.Now(),
		Imports: make([]*Lock, len(ds)),
	}

	for i := 0; i < len(ds); i++ {
		lf.Imports[i] = &Lock{
			Name:    ds[i].Name,
			Version: ds[i].Pin,
		}
	}

	sort.Sort(lf.Imports)

	return lf
}

func LockfileFromMap(ds map[string]*Dependency, hash string) *Lockfile {
	lf := &Lockfile{
		Hash:    hash,
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

	sort.Sort(lf.Imports)

	return lf
}
