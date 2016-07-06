package cfg

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/sdboyer/vsolver"

	"gopkg.in/yaml.v2"
)

// Lockfile represents a glide.lock file.
type Lockfile struct {
	Hash       string    `yaml:"hash"`
	Updated    time.Time `yaml:"updated"`
	Imports    Locks     `yaml:"imports"`
	DevImports Locks     `yaml:"testImports"`
}

// LockfileFromSolverLock transforms a vsolver.Lock into a glide *Lockfile.
func LockfileFromSolverLock(r vsolver.Lock) *Lockfile {
	if r == nil {
		return nil
	}

	// Create and write out a new lock file from the result
	lf := &Lockfile{
		Hash:    hex.EncodeToString(r.InputHash()),
		Updated: time.Now(),
	}

	for _, p := range r.Projects() {
		pi := p.Ident()
		l := &Lock{
			Name:    string(pi.LocalName),
			VcsType: "", // TODO allow this to be extracted from sm
		}

		if l.Name != pi.NetworkName && pi.NetworkName != "" {
			l.Repository = pi.NetworkName
		}

		v := p.Version()
		if pv, ok := v.(vsolver.PairedVersion); ok {
			l.Version = pv.Underlying().String()
		} else {
			l.Version = v.String()
		}

		lf.Imports = append(lf.Imports, l)
	}

	return lf
}

// LockfileFromYaml returns an instance of Lockfile from YAML
func LockfileFromYaml(yml []byte) (*Lockfile, error) {
	lock := &Lockfile{}
	err := yaml.Unmarshal([]byte(yml), &lock)
	return lock, err
}

// Marshal converts a Config instance to YAML
func (lf *Lockfile) Marshal() ([]byte, error) {
	sort.Sort(lf.Imports)
	sort.Sort(lf.DevImports)
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

// InputHash returns the hash of the input arguments that resulted in this lock
// file.
func (lf *Lockfile) InputHash() []byte {
	b, err := hex.DecodeString(lf.Hash)
	if err != nil {
		return nil
	}
	return b
}

// Projects returns the list of projects enumerated in the lock file.
func (lf *Lockfile) Projects() []vsolver.LockedProject {
	all := append(lf.Imports, lf.DevImports...)
	lp := make([]vsolver.LockedProject, len(all))

	for k, l := range all {
		// TODO guess the version type. ugh
		var v vsolver.Version

		// semver first
		_, err := semver.NewVersion(l.Version)
		if err == nil {
			v = vsolver.NewVersion(l.Version)
		} else {
			// Crappy heuristic to cover hg and git, but not bzr. Or (lol) svn
			if len(l.Version) == 40 {
				v = vsolver.Revision(l.Version)
			} else {
				// Otherwise, assume it's a branch
				v = vsolver.NewBranch(l.Version)
			}
		}

		lp[k] = vsolver.NewLockedProject(vsolver.ProjectName(l.Name), v, l.Repository, l.Name, nil)
	}

	return lp
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
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Repository  string   `yaml:"repo,omitempty"`
	VcsType     string   `yaml:"vcs,omitempty"`
	Subpackages []string `yaml:"subpackages,omitempty"`
	Arch        []string `yaml:"arch,omitempty"`
	Os          []string `yaml:"os,omitempty"`
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
	}
}

// NewLockfile is used to create an instance of Lockfile.
func NewLockfile(ds, tds Dependencies, hash string) *Lockfile {
	lf := &Lockfile{
		Hash:       hash,
		Updated:    time.Now(),
		Imports:    make([]*Lock, len(ds)),
		DevImports: make([]*Lock, len(tds)),
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

	sort.Sort(lf.Imports)

	for i := 0; i < len(tds); i++ {
		lf.DevImports[i] = &Lock{
			Name:        tds[i].Name,
			Version:     tds[i].Pin,
			Repository:  tds[i].Repository,
			VcsType:     tds[i].VcsType,
			Subpackages: tds[i].Subpackages,
			Arch:        tds[i].Arch,
			Os:          tds[i].Os,
		}
	}

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
