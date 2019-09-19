package cfg

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/glide/util"
	"gopkg.in/yaml.v2"
)

// Lockfile represents a glide.lock file.
type Lockfile struct {
	Hash       string    `yaml:"hash"`
	Updated    time.Time `yaml:"updated"`
	Imports    Locks     `yaml:"imports"`
	DevImports Locks     `yaml:"testImports"`
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

// MarshalYAML is a hook for gopkg.in/yaml.v2.
// It sorts import subpackages lexicographically for reproducibility.
func (lf *Lockfile) MarshalYAML() (interface{}, error) {
	for _, imp := range lf.Imports {
		sort.Strings(imp.Subpackages)
	}

	// Ensure elements on testImport don't already exist on import.
	var newDI Locks
	var found bool
	for _, imp := range lf.DevImports {
		found = false
		for i := 0; i < len(lf.Imports); i++ {
			if lf.Imports[i].Name == imp.Name {
				found = true
				if lf.Imports[i].Version != imp.Version {
					return lf, fmt.Errorf("Generating lock YAML produced conflicting versions of %s. import (%s), testImport (%s)", imp.Name, lf.Imports[i].Version, imp.Version)
				}
			}
		}

		if !found {
			newDI = append(newDI, imp)
		}
	}
	lf.DevImports = newDI

	for _, imp := range lf.DevImports {
		sort.Strings(imp.Subpackages)
	}
	return lf, nil
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
	for _, l := range lf.Imports {
		o = bytes.Replace(o, []byte("version: "+l.Version), []byte("version: "+l.Version+" # "+l.Dep.Original+", "+util.Date(l.Dep.CommitInfo.Date)), 1)
	}
	for _, l := range lf.DevImports {
		o = bytes.Replace(o, []byte("version: "+l.Version), []byte("version: "+l.Version+" # "+l.Dep.Original+", "+util.Date(l.Dep.CommitInfo.Date)), 1)
	}
	return ioutil.WriteFile(lockpath, o, 0666)
}

// Clone returns a clone of Lockfile
func (lf *Lockfile) Clone() *Lockfile {
	n := &Lockfile{}
	n.Hash = lf.Hash
	n.Updated = lf.Updated
	n.Imports = lf.Imports.Clone()
	n.DevImports = lf.DevImports.Clone()

	return n
}

// Fingerprint returns a hash of the contents minus the date. This allows for
// two lockfiles to be compared irrespective of their updated times.
func (lf *Lockfile) Fingerprint() ([32]byte, error) {
	c := lf.Clone()
	c.Updated = time.Time{} // Set the time to be the nil equivalent
	sort.Sort(c.Imports)
	sort.Sort(c.DevImports)
	yml, err := c.Marshal()
	if err != nil {
		return [32]byte{}, err
	}

	return sha256.Sum256(yml), nil
}

// ReadLockFile loads the contents of a glide.lock file.
func ReadLockFile(lockpath string) (*Lockfile, error) {
	yml, err := ioutil.ReadFile(lockpath)
	if err != nil {
		return nil, err
	}
	lock, err := LockfileFromYaml(yml)
	if err != nil {
		return nil, err
	}
	return lock, nil
}

// Locks is a slice of locked dependencies.
type Locks []*Lock

// Clone returns a Clone of Locks.
func (l Locks) Clone() Locks {
	n := make(Locks, 0, len(l))
	for _, v := range l {
		n = append(n, v.Clone())
	}
	return n
}

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
	Name        string      `yaml:"name"`
	Version     string      `yaml:"version"`
	Repository  string      `yaml:"repo,omitempty"`
	VcsType     string      `yaml:"vcs,omitempty"`
	Subpackages []string    `yaml:"subpackages,omitempty"`
	Arch        []string    `yaml:"arch,omitempty"`
	Os          []string    `yaml:"os,omitempty"`
	Dep         *Dependency `yaml:"-"`
}

// Clone creates a clone of a Lock.
func (l *Lock) Clone() *Lock {
	return &Lock{
		Name:        l.Name,
		Version:     l.Version,
		Repository:  l.Repository,
		VcsType:     l.VcsType,
		Subpackages: l.Subpackages,
		Arch:        l.Arch,
		Os:          l.Os,
		Dep:         l.Dep,
	}
}

// LockFromDependency converts a Dependency to a Lock
func LockFromDependency(dep *Dependency) *Lock {
	return &Lock{
		Name:        dep.Name,
		Version:     dep.Pin,
		Repository:  dep.Repository,
		VcsType:     dep.VcsType,
		Subpackages: dep.Subpackages,
		Arch:        dep.Arch,
		Os:          dep.Os,
		Dep:         dep,
	}
}

// NewLockfile is used to create an instance of Lockfile.
func NewLockfile(ds, tds Dependencies, hash string) (*Lockfile, error) {
	lf := &Lockfile{
		Hash:       hash,
		Updated:    time.Now(),
		Imports:    make([]*Lock, len(ds)),
		DevImports: make([]*Lock, 0),
	}

	for i := 0; i < len(ds); i++ {
		lf.Imports[i] = LockFromDependency(ds[i])
	}

	sort.Sort(lf.Imports)

	var found bool
	for i := 0; i < len(tds); i++ {
		found = false
		for ii := 0; ii < len(ds); ii++ {
			if ds[ii].Name == tds[i].Name {
				found = true
				if ds[ii].Reference != tds[i].Reference {
					return &Lockfile{}, fmt.Errorf("Generating lock produced conflicting versions of %s. import (%s), testImport (%s)", tds[i].Name, ds[ii].Reference, tds[i].Reference)
				}
				break
			}
		}
		if !found {
			lf.DevImports = append(lf.DevImports, LockFromDependency(tds[i]))
		}
	}

	sort.Sort(lf.DevImports)

	return lf, nil
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
		lf.Imports[i] = LockFromDependency(dep)
		lf.Imports[i].Name = name
		i++
	}

	sort.Sort(lf.Imports)

	return lf
}
