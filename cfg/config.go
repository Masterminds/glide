package cfg

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"

	"github.com/Masterminds/glide/mirrors"
	"github.com/Masterminds/vcs"
	"github.com/sdboyer/gps"
	"gopkg.in/yaml.v2"
)

// Config is the top-level configuration object.
type Config struct {

	// Name is the name of the package or application.
	Name string `yaml:"package"`

	// Description is a short description for a package, application, or library.
	// This description is similar but different to a Go package description as
	// it is for marketing and presentation purposes rather than technical ones.
	Description string `json:"description,omitempty"`

	// Home is a url to a website for the package.
	Home string `yaml:"homepage,omitempty"`

	// License provides either a SPDX license or a path to a file containing
	// the license. For more information on SPDX see http://spdx.org/licenses/.
	// When more than one license an SPDX expression can be used.
	License string `yaml:"license,omitempty"`

	// Owners is an array of owners for a project. See the Owner type for
	// more detail. These can be one or more people, companies, or other
	// organizations.
	Owners Owners `yaml:"owners,omitempty"`

	// Ignore contains a list of packages to ignore fetching. This is useful
	// when walking the package tree (including packages of packages) to list
	// those to skip.
	Ignore []string `yaml:"ignore,omitempty"`

	// Imports contains a list of all dependency constraints for a project. For
	// more detail on how these are captured see the Dependency type.
	// TODO rename
	// TODO mapify
	Imports Dependencies `yaml:"dependencies"`

	// DevImports contains the test or other development dependency constraints
	// for a project. See the Dependency type for more details on how this is
	// recorded.
	// TODO rename
	// TODO mapify
	DevImports Dependencies `yaml:"testDependencies"`
}

// A transitive representation of a dependency for importing and exporting to yaml.
type cf struct {
	Name        string       `yaml:"package"`
	Description string       `yaml:"description,omitempty"`
	Home        string       `yaml:"homepage,omitempty"`
	License     string       `yaml:"license,omitempty"`
	Owners      Owners       `yaml:"owners,omitempty"`
	Ignore      []string     `yaml:"ignore,omitempty"`
	Imports     Dependencies `yaml:"dependencies,omitempty"`
	DevImports  Dependencies `yaml:"testDependencies,omitempty"`
	// these fields guarantee that this struct fails to unmarshal legacy yamls
	Compat  int `yaml:"import,omitempty"`
	Compat2 int `yaml:"testImport,omitempty"`
}

// ConfigFromYaml returns an instance of Config from YAML
func ConfigFromYaml(yml []byte) (cfg *Config, legacy bool, err error) {
	cfg = &Config{}
	err = yaml.Unmarshal(yml, cfg)
	if err != nil {
		lcfg := &lConfig1{}
		err = yaml.Unmarshal(yml, &lcfg)
		if err == nil {
			legacy = true
			cfg, err = lcfg.Convert()
		}
	}

	return
}

// Marshal converts a Config instance to YAML
func (c *Config) Marshal() ([]byte, error) {
	yml, err := yaml.Marshal(&c)
	if err != nil {
		return []byte{}, err
	}
	return yml, nil
}

// UnmarshalYAML is a hook for gopkg.in/yaml.v2 in the unmarshalling process
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	newConfig := &cf{}
	if err := unmarshal(&newConfig); err != nil {
		return err
	}
	c.Name = newConfig.Name
	c.Description = newConfig.Description
	c.Home = newConfig.Home
	c.License = newConfig.License
	c.Owners = newConfig.Owners
	c.Ignore = newConfig.Ignore
	c.Imports = newConfig.Imports
	c.DevImports = newConfig.DevImports

	// Cleanup the Config object now that we have it.
	err := c.DeDupe()

	return err
}

// MarshalYAML is a hook for gopkg.in/yaml.v2 in the marshaling process
func (c *Config) MarshalYAML() (interface{}, error) {
	newConfig := &cf{
		Name:        c.Name,
		Description: c.Description,
		Home:        c.Home,
		License:     c.License,
		Owners:      c.Owners,
		Ignore:      c.Ignore,
	}
	i, err := c.Imports.Clone().DeDupe()
	if err != nil {
		return newConfig, err
	}

	di, err := c.DevImports.Clone().DeDupe()
	if err != nil {
		return newConfig, err
	}

	newConfig.Imports = i
	newConfig.DevImports = di

	return newConfig, nil
}

// HasDependency returns true if the given name is listed as an import or dev import.
func (c *Config) HasDependency(name string) bool {
	for _, d := range c.Imports {
		if d.Name == name {
			return true
		}
	}
	for _, d := range c.DevImports {
		if d.Name == name {
			return true
		}
	}
	return false
}

// DependencyConstraints lists all the non-test dependency constraints
// described in a glide manifest in a way gps will understand.
func (c *Config) DependencyConstraints() gps.ProjectConstraints {
	return gpsifyDeps(c.Imports)
}

// TestDependencyConstraints lists all the test dependency constraints described
// in a glide manifest in a way gps will understand.
func (c *Config) TestDependencyConstraints() gps.ProjectConstraints {
	return gpsifyDeps(c.DevImports)
}

func gpsifyDeps(deps Dependencies) gps.ProjectConstraints {
	cp := make(gps.ProjectConstraints, len(deps))
	for _, d := range deps {
		cp[gps.ProjectRoot(d.Name)] = gps.ProjectProperties{
			NetworkName: d.Repository,
			Constraint:  d.GetConstraint(),
		}
	}

	return cp
}

func (c *Config) IgnorePackages() map[string]bool {
	m := make(map[string]bool)
	for _, ig := range c.Ignore {
		m[ig] = true
	}
	return m
}

func (c *Config) Overrides() gps.ProjectConstraints {
	return nil
}

// HasIgnore returns true if the given name is listed on the ignore list.
func (c *Config) HasIgnore(name string) bool {
	for _, v := range c.Ignore {

		// Check for both a name and to make sure sub-packages are ignored as
		// well.
		if v == name || strings.HasPrefix(name, v+"/") {
			return true
		}
	}

	return false
}

// Clone performs a deep clone of the Config instance
func (c *Config) Clone() *Config {
	n := &Config{}
	n.Name = c.Name
	n.Description = c.Description
	n.Home = c.Home
	n.License = c.License
	n.Owners = c.Owners.Clone()
	n.Ignore = c.Ignore
	n.Imports = c.Imports.Clone()
	n.DevImports = c.DevImports.Clone()
	return n
}

// WriteFile writes a Glide YAML file.
//
// This is a convenience function that marshals the YAML and then writes it to
// the given file. If the file exists, it will be clobbered.
func (c *Config) WriteFile(glidepath string) error {
	o, err := c.Marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(glidepath, o, 0666)
}

// DeDupe consolidates duplicate dependencies on a Config instance
func (c *Config) DeDupe() error {
	// Remove duplicates in the imports
	var err error
	c.Imports, err = c.Imports.DeDupe()
	if err != nil {
		return err
	}
	c.DevImports, err = c.DevImports.DeDupe()
	if err != nil {
		return err
	}

	// If the name on the config object is part of the imports remove it.
	found := -1
	for i, dep := range c.Imports {
		if dep.Name == c.Name {
			found = i
		}
	}
	if found >= 0 {
		c.Imports = append(c.Imports[:found], c.Imports[found+1:]...)
	}

	found = -1
	for i, dep := range c.DevImports {
		if dep.Name == c.Name {
			found = i
		}
	}
	if found >= 0 {
		c.DevImports = append(c.DevImports[:found], c.DevImports[found+1:]...)
	}

	// If something is on the ignore list remove it from the imports.
	for _, v := range c.Ignore {
		found = -1
		for k, d := range c.Imports {
			if v == d.Name {
				found = k
			}
		}
		if found >= 0 {
			c.Imports = append(c.Imports[:found], c.Imports[found+1:]...)
		}

		found = -1
		for k, d := range c.DevImports {
			if v == d.Name {
				found = k
			}
		}
		if found >= 0 {
			c.DevImports = append(c.DevImports[:found], c.DevImports[found+1:]...)
		}
	}

	return nil
}

// AddImport appends dependencies to the import list, deduplicating as we go.
func (c *Config) AddImport(deps ...*Dependency) error {
	t := c.Imports
	t = append(t, deps...)
	t, err := t.DeDupe()
	if err != nil {
		return err
	}
	c.Imports = t
	return nil
}

// Hash generates a sha256 hash for a given Config
func (c *Config) Hash() (string, error) {
	yml, err := c.Marshal()
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write(yml)
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// Dependencies is a collection of Dependency
type Dependencies []*Dependency

// Get a dependency by name
func (d Dependencies) Get(name string) *Dependency {
	for _, dep := range d {
		if dep.Name == name {
			return dep
		}
	}
	return nil
}

// Has checks if a dependency is on a list of dependencies such as import or testImport
func (d Dependencies) Has(name string) bool {
	for _, dep := range d {
		if dep.Name == name {
			return true
		}
	}
	return false
}

// Remove removes a dependency from a list of dependencies
func (d Dependencies) Remove(name string) Dependencies {
	found := -1
	for i, dep := range d {
		if dep.Name == name {
			found = i
		}
	}

	if found >= 0 {
		copy(d[found:], d[found+1:])
		d[len(d)-1] = nil
		return d[:len(d)-1]
	}
	return d
}

// Clone performs a deep clone of Dependencies
func (d Dependencies) Clone() Dependencies {
	n := make(Dependencies, 0, len(d))
	for _, v := range d {
		n = append(n, v.Clone())
	}
	return n
}

// DeDupe cleans up duplicates on a list of dependencies.
func (d Dependencies) DeDupe() (Dependencies, error) {
	checked := map[string]int{}
	imports := make(Dependencies, 0, 1)
	i := 0
	for _, dep := range d {
		// The first time we encounter a dependency add it to the list
		if val, ok := checked[dep.Name]; !ok {
			checked[dep.Name] = i
			imports = append(imports, dep)
			i++
		} else {
			// In here we've encountered a dependency for the second time.
			// Make sure the details are the same or return an error.
			v := imports[val]
			// Have to do string-based comparison
			if dep.ConstraintsEq(*v) {
				return d, fmt.Errorf("Import %s repeated with different versions '%s' and '%s'", dep.Name, dep.GetConstraint(), v.GetConstraint())
			}
			if dep.Repository != v.Repository {
				return d, fmt.Errorf("Import %s repeated with different Repository details", dep.Name)
			}
		}
	}

	return imports, nil
}

// Dependency describes a package that the present package depends upon.
type Dependency struct {
	Name       string
	VcsType    string // TODO remove
	Repository string
	Branch     string
	Version    string
}

// A transitive representation of a dependency for yaml import/export.
type dep struct {
	Name       string `yaml:"package"`
	Version    string `yaml:"version,omitempty"`
	Branch     string `yaml:"branch,omitempty"`
	Repository string `yaml:"repo,omitempty"`
}

// DependencyFromLock converts a Lock to a Dependency
func DependencyFromLock(lock *Lock) *Dependency {
	d := &Dependency{
		Name:       lock.Name,
		Repository: lock.Repository,
	}

	// Because it's not allowed to have both, if we see both, prefer version
	// over branch
	if lock.Version != "" {
		d.Version = lock.Version
	} else if lock.Branch != "" {
		d.Branch = lock.Branch
	} else {
		d.Version = lock.Revision
	}

	return d
}

// GetConstraint constructs an appropriate gps.Constraint from the Dependency's
// string input data.
func (d Dependency) GetConstraint() gps.Constraint {
	// If neither or both Version and Branch are set, accept anything
	if d.IsUnconstrained() {
		return gps.Any()
	} else if d.Version != "" {
		return DeduceConstraint(d.Version)
	} else {
		// only case left is a non-empty branch
		return gps.NewBranch(d.Branch)
	}
}

// IsUnconstrained indicates if this dependency has no constraint information,
// version or branch.
func (d Dependency) IsUnconstrained() bool {
	return (d.Version != "" && d.Branch != "") || (d.Version == "" && d.Branch == "")
}

// ConstraintsEq checks if the constraints on two Dependency are exactly equal.
func (d Dependency) ConstraintsEq(d2 Dependency) bool {
	// Having both branch and version set is always an error, so if either have
	// it, then return false
	if (d.Version != "" && d.Branch != "") || (d2.Version != "" && d2.Branch != "") {
		return false
	}
	// Neither being set, though, is OK
	if (d.Version == "" && d.Branch == "") || (d2.Version == "" && d2.Branch == "") {
		return true
	}

	// Now, xors
	if d.Version != "" && d.Version == d2.Version {
		return true
	}
	if d.Branch == d2.Branch {
		return true
	}
	return false
}

// UnmarshalYAML is a hook for gopkg.in/yaml.v2 in the unmarshaling process
func (d *Dependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	newDep := dep{}
	err := unmarshal(&newDep)
	if err != nil {
		return err
	}

	if newDep.Version != "" && newDep.Branch != "" {
		return fmt.Errorf("Cannot set both a both a branch and a version constraint for %q", d.Name)
	}

	d.Name = newDep.Name
	d.Repository = newDep.Repository
	d.Version = newDep.Version
	d.Branch = newDep.Branch

	return nil
}

// DeduceConstraint tries to puzzle out what kind of version is given in a string -
// semver, a revision, or as a fallback, a plain tag
func DeduceConstraint(s string) gps.Constraint {
	// always semver if we can
	c, err := gps.NewSemverConstraint(s)
	if err == nil {
		return c
	}

	slen := len(s)
	if slen == 40 {
		if _, err = hex.DecodeString(s); err == nil {
			// Whether or not it's intended to be a SHA1 digest, this is a
			// valid byte sequence for that, so go with Revision. This
			// covers git and hg
			return gps.Revision(s)
		}
	}
	// Next, try for bzr, which has a three-component GUID separated by
	// dashes. There should be two, but the email part could contain
	// internal dashes
	if strings.Count(s, "-") >= 2 {
		// Work from the back to avoid potential confusion from the email
		i3 := strings.LastIndex(s, "-")
		// Skip if - is last char, otherwise this would panic on bounds err
		if slen == i3+1 {
			return gps.NewVersion(s)
		}

		if _, err = hex.DecodeString(s[i3+1:]); err == nil {
			i2 := strings.LastIndex(s[:i3], "-")
			if _, err = strconv.ParseUint(s[i2+1:i3], 10, 64); err == nil {
				// Getting this far means it'd pretty much be nuts if it's not a
				// bzr rev, so don't bother parsing the email.
				return gps.Revision(s)
			}
		}
	}

	// If not a plain SHA1 or bzr custom GUID, assume a plain version.
	//
	// svn, you ask? lol, madame. lol.
	return gps.NewVersion(s)
}

// MarshalYAML is a hook for gopkg.in/yaml.v2 in the marshaling process
func (d *Dependency) MarshalYAML() (interface{}, error) {
	newDep := &dep{
		Name:       d.Name,
		Repository: d.Repository,
		Version:    d.Version,
		Branch:     d.Branch,
	}

	return newDep, nil
}

// Remote returns the remote location to fetch source from. This location is
// the central place where mirrors can alter the location.
func (d *Dependency) Remote() string {
	var r string

	if d.Repository != "" {
		r = d.Repository
	} else {
		r = "https://" + d.Name
	}

	f, nr, _ := mirrors.Get(r)
	if f {
		return nr
	}

	return r
}

// Vcs returns the VCS type to fetch source from.
func (d *Dependency) Vcs() string {
	var r string

	if d.Repository != "" {
		r = d.Repository
	} else {
		r = "https://" + d.Name
	}

	f, _, nv := mirrors.Get(r)
	if f {
		return nv
	}

	return d.VcsType
}

// GetRepo retrieves a Masterminds/vcs repo object configured for the root
// of the package being retrieved.
// TODO remove
func (d *Dependency) GetRepo(dest string) (vcs.Repo, error) {
	// The remote location is either the configured repo or the package
	// name as an https url.
	remote := d.Remote()

	VcsType := d.Vcs()

	// If the VCS type has a value we try that first.
	if len(VcsType) > 0 && VcsType != "None" {
		switch vcs.Type(VcsType) {
		case vcs.Git:
			return vcs.NewGitRepo(remote, dest)
		case vcs.Svn:
			return vcs.NewSvnRepo(remote, dest)
		case vcs.Hg:
			return vcs.NewHgRepo(remote, dest)
		case vcs.Bzr:
			return vcs.NewBzrRepo(remote, dest)
		default:
			return nil, fmt.Errorf("Unknown VCS type %s set for %s", VcsType, d.Name)
		}
	}

	// When no type set we try to autodetect.
	return vcs.NewRepo(remote, dest)
}

// Clone creates a clone of a Dependency
func (d *Dependency) Clone() *Dependency {
	var d2 Dependency
	d2 = *d
	return &d2
}

// Owners is a list of owners for a project.
type Owners []*Owner

// Clone performs a deep clone of Owners
func (o Owners) Clone() Owners {
	n := make(Owners, 0, 1)
	for _, v := range o {
		n = append(n, v.Clone())
	}
	return n
}

// Owner describes an owner of a package. This can be a person, company, or
// other organization. This is useful if someone needs to contact the
// owner of a package to address things like a security issue.
type Owner struct {

	// Name describes the name of an organization.
	Name string `yaml:"name,omitempty"`

	// Email is an email address to reach the owner at.
	Email string `yaml:"email,omitempty"`

	// Home is a url to a website for the owner.
	Home string `yaml:"homepage,omitempty"`
}

// Clone creates a clone of a Dependency
func (o *Owner) Clone() *Owner {
	return &Owner{
		Name:  o.Name,
		Email: o.Email,
		Home:  o.Home,
	}
}

func stringArrayDeDupe(s []string, items ...string) []string {
	for _, item := range items {
		exists := false
		for _, v := range s {
			if v == item {
				exists = true
			}
		}
		if !exists {
			s = append(s, item)
		}
	}
	sort.Strings(s)
	return s
}

func filterVcsType(vcs string) string {
	switch vcs {
	case "git", "hg", "bzr", "svn":
		return vcs
	case "mercurial":
		return "hg"
	case "bazaar":
		return "bzr"
	case "subversion":
		return "svn"
	default:
		return ""
	}
}

func normalizeSlash(k string) string {
	return strings.Replace(k, "\\", "/", -1)
}
