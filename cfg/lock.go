package cfg

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"github.com/sdboyer/gps"

	"gopkg.in/yaml.v2"
)

// Lockfile represents a glide.lock file.
type Lockfile struct {
	Hash       string    `yaml:"hash"`
	Updated    time.Time `yaml:"updated"`
	Imports    Locks     `yaml:"imports"`
	DevImports Locks     `yaml:"testImports"` // TODO remove and fold in as prop
}

// LockfileFromSolverLock transforms a gps.Lock into a glide *Lockfile.
func LockfileFromSolverLock(r gps.Lock) (*Lockfile, error) {
	if r == nil {
		return nil, fmt.Errorf("no gps lock data provided to transform")
	}

	// Create and write out a new lock file from the result
	lf := &Lockfile{
		Hash:    hex.EncodeToString(r.InputHash()),
		Updated: time.Now(),
	}

	for _, p := range r.Projects() {
		pi := p.Ident()
		l := &Lock{
			Name: string(pi.ProjectRoot),
		}

		if l.Name != pi.NetworkName && pi.NetworkName != "" {
			l.Repository = pi.NetworkName
		}

		v := p.Version()
		// There's (currently) no way gps can emit a non-paired version in a
		// solution, so this unchecked type assertion should be safe.
		//
		// TODO might still be better to check and return out with an err if
		// not, though
		switch tv := v.(type) {
		case gps.Revision:
			l.Revision = tv.String()
		case gps.PairedVersion:
			l.Revision = v.(gps.PairedVersion).Underlying().String()
			switch v.Type() {
			case "branch":
				l.Branch = v.String()
			case "semver", "version":
				l.Version = v.String()
			}
		case gps.UnpairedVersion:
			// this should not be possible - error if we hit it
			return nil, fmt.Errorf("should not be possible - gps returned an unpaired version for %s", pi)
		}

		lf.Imports = append(lf.Imports, l)
	}

	return lf, nil
}

// LockfileFromYaml returns an instance of Lockfile from YAML
func LockfileFromYaml(yml []byte) (*Lockfile, bool, error) {
	lock := &Lockfile{}
	err := yaml.Unmarshal([]byte(yml), lock)
	if err == nil {
		return lock, false, nil
	}

	llock := &lLockfile1{}
	err2 := yaml.Unmarshal([]byte(yml), llock)
	if err2 != nil {
		return nil, false, err2
	}
	return llock.Convert(), true, nil
}

// Marshal converts a Lockfile instance to YAML
func (lf *Lockfile) Marshal() ([]byte, error) {
	sort.Sort(lf.Imports)
	sort.Sort(lf.DevImports)
	yml, err := yaml.Marshal(&lf)
	if err != nil {
		return []byte{}, err
	}
	return yml, nil
}

// MarshalYAML is a hook for gopkg.in/yaml.v2.
// It sorts import subpackages lexicographically for reproducibility.
func (lf *Lockfile) MarshalYAML() (interface{}, error) {
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
func (lf *Lockfile) Projects() []gps.LockedProject {
	all := append(lf.Imports, lf.DevImports...)
	lp := make([]gps.LockedProject, len(all))

	for k, l := range all {
		r := gps.Revision(l.Revision)

		var v gps.Version
		if l.Version != "" {
			v = gps.NewVersion(l.Version).Is(r)
		} else if l.Branch != "" {
			v = gps.NewBranch(l.Branch).Is(r)
		} else {
			v = r
		}

		id := gps.ProjectIdentifier{
			ProjectRoot: gps.ProjectRoot(l.Name),
			NetworkName: l.Repository,
		}
		lp[k] = gps.NewLockedProject(id, v, nil)
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
// TODO remove, or seriously re-adapt
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
	lock, _, err := LockfileFromYaml(yml)
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
	Name       string `yaml:"name"`
	Version    string `yaml:"version,omitempty"`
	Branch     string `yaml:"branch,omitempty"`
	Revision   string `yaml:"revision"`
	Repository string `yaml:"repo,omitempty"`
}

func (l *Lock) UnmarshalYAML(unmarshal func(interface{}) error) error {
	nl := struct {
		Name       string `yaml:"name"`
		Version    string `yaml:"version,omitempty"`
		Branch     string `yaml:"branch,omitempty"`
		Revision   string `yaml:"revision"`
		Repository string `yaml:"repo,omitempty"`
	}{}

	err := unmarshal(&nl)
	if err != nil {
		return err
	}

	// If Revision field is empty, then we can be certain this is either a
	// legacy file, or just plain invalid
	if nl.Revision == "" {
		return fmt.Errorf("dependency %s is missing a revision; is this a legacy glide.lock file?", nl.Name)
	}

	l.Name = nl.Name
	l.Version = nl.Version
	l.Branch = nl.Branch
	l.Revision = nl.Revision
	l.Repository = nl.Repository

	return nil
}

// Clone creates a clone of a Lock.
func (l *Lock) Clone() *Lock {
	var l2 Lock
	l2 = *l
	return &l2
}

// LockFromDependency converts a Dependency to a Lock
// TODO remove
func LockFromDependency(dep *Dependency) *Lock {
	l := &Lock{
		Name:       dep.Name,
		Repository: dep.Repository,
	}

	return l
}

// NewLockfile is used to create an instance of Lockfile.
// TODO remove
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
				if ds[ii].ConstraintsEq(*tds[i]) {
					return &Lockfile{}, fmt.Errorf("Generating lock produced conflicting versions of %s. import (%s), testImport (%s)", tds[i].Name, ds[ii].GetConstraint(), tds[i].GetConstraint())
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
// TODO remove
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
