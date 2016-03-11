package cfg

import (
	"io/ioutil"
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

// LockfileFromYaml returns an instance of Lockfile from YAML
func LockfileFromYaml(yml []byte) (*Lockfile, error) {
	lock := &Lockfile{}
	err := yaml.Unmarshal([]byte(yml), &lock)
	return lock, err
}

// Marshal converts a Config instance to YAML
func (lf *Lockfile) Marshal() ([]byte, error) {
	yml, err := yaml.Marshal(&lf)
	if err != nil {
		return []byte{}, err
	}
	return yml, nil
}

// WriteFile writes a Glide lock file.
//
// This is a convenience function that marshals the YAML and then writes it to
// the given file. If the file exists, it will be clobbered.
func (lf *Lockfile) WriteFile(lockpath string) error {
	o, err := lf.Marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(lockpath, o, 0666)
}

// Locks is a slice of locked dependencies.
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

// Lock represents an individual locked dependency.
type Lock struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Repository  string   `yaml:"repo,omitempty"`
	VcsType     string   `yaml:"vcs,omitempty"`
	Subpackages []string `yaml:"subpackages,omitempty"`
	Arch        []string `yaml:"arch,omitempty"`
	Os          []string `yaml:"os,omitempty"`
}

// NewLockfile is used to create an instance of Lockfile.
func NewLockfile(ds Dependencies, devs Dependencies, hash string) *Lockfile {
	lf := &Lockfile{
		Hash:       hash,
		Updated:    time.Now(),
		Imports:    make([]*Lock, len(ds)),
		DevImports: make([]*Lock, len(devs)),
	}

	for i := 0; i < len(ds); i++ {
		lf.Imports[i] = &Lock{
			Name:        ds[i].Name,
			Version:     ds[i].Pin,
			Repository:  ds[i].Repository,
			VcsType:     ds[i].VcsType,
			Subpackages: ds[i].Subpackages,
			Arch:        ds[i].Arch,
			Os:          ds[i].Os,
		}
	}

	for i := 0; i < len(devs); i++ {
		lf.DevImports[i] = &Lock{
			Name:        devs[i].Name,
			Version:     devs[i].Pin,
			Repository:  devs[i].Repository,
			VcsType:     devs[i].VcsType,
			Subpackages: devs[i].Subpackages,
			Arch:        devs[i].Arch,
			Os:          devs[i].Os,
		}
	}

	sort.Sort(lf.Imports)
	sort.Sort(lf.DevImports)

	return lf
}

// LockfileFromMap takes a map of dependencies and generates a lock Lockfile instance.
func LockfileFromMap(ds map[string]*Dependency, hash string) *Lockfile {
	lf := &Lockfile{
		Hash:    hash,
		Updated: time.Now(),
		Imports: make([]*Lock, len(ds)),
	}

	i := 0
	for name, dep := range ds {
		lf.Imports[i] = &Lock{
			Name:        name,
			Version:     dep.Pin,
			Repository:  dep.Repository,
			VcsType:     dep.VcsType,
			Subpackages: dep.Subpackages,
			Arch:        dep.Arch,
			Os:          dep.Os,
		}
		i++
	}

	sort.Sort(lf.Imports)

	return lf
}
