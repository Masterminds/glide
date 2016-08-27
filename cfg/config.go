package cfg

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"

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
	Imports Dependencies `yaml:"dependencies"`

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
	Imports     Dependencies `yaml:"dependencies"`
	DevImports  Dependencies `yaml:"testDependencies,omitempty"`
}

// ConfigFromYaml returns an instance of Config from YAML
func ConfigFromYaml(yml []byte) (*Config, bool, error) {
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(yml), &cfg)
	if err != nil {
		lcfg := &lConfig1{}
		err = yaml.Unmarshal([]byte(yml), &lcfg)
		if err != nil {
			// TODO(sdboyer) convert to new form, then return
		}
	}
	return cfg, false, err
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
func (c *Config) DependencyConstraints() []gps.ProjectConstraint {
	return depsToVSolver(c.Imports)
}

// TestDependencyConstraints lists all the test dependency constraints described
// in a glide manifest in a way gps will understand.
func (c *Config) TestDependencyConstraints() []gps.ProjectConstraint {
	return depsToVSolver(c.DevImports)
}

func depsToVSolver(deps Dependencies) []gps.ProjectConstraint {
	cp := make([]gps.ProjectConstraint, len(deps))
	for k, d := range deps {
		cp[k] = gps.ProjectConstraint{
			Ident: gps.ProjectIdentifier{
				ProjectRoot: gps.ProjectRoot(d.Name),
				NetworkName: d.Repository,
			},
			Constraint: d.Constraint,
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
			if dep.Constraint.String() != v.Constraint.String() {
				return d, fmt.Errorf("Import %s repeated with different versions '%s' and '%s'", dep.Name, dep.Constraint, v.Constraint)
			}
			if dep.Repository != v.Repository {
				return d, fmt.Errorf("Import %s repeated with different Repository details", dep.Name)
			}
			if !reflect.DeepEqual(dep.Os, v.Os) || !reflect.DeepEqual(dep.Arch, v.Arch) {
				return d, fmt.Errorf("Import %s repeated with different OS or Architecture filtering", dep.Name)
			}
		}
	}

	return imports, nil
}

// Dependency describes a package that the present package depends upon.
type Dependency struct {
	Name       string
	VcsType    string // TODO remove
	Constraint gps.Constraint
	Repository string
	Arch       []string
	Os         []string
}

// A transitive representation of a dependency for yaml import/export.
type dep struct {
	Name       string   `yaml:"package"`
	Reference  string   `yaml:"version,omitempty"`
	Branch     string   `yaml:"branch,omitempty"`
	Repository string   `yaml:"repo,omitempty"`
	Arch       []string `yaml:"arch,omitempty"`
	Os         []string `yaml:"os,omitempty"`
}

// DependencyFromLock converts a Lock to a Dependency
func DependencyFromLock(lock *Lock) *Dependency {
	d := &Dependency{
		Name:       lock.Name,
		Repository: lock.Repository,
	}

	r := gps.Revision(lock.Revision)
	if lock.Version != "" {
		d.Constraint = gps.NewVersion(lock.Version).Is(r)
	} else if lock.Branch != "" {
		d.Constraint = gps.NewBranch(lock.Version).Is(r)
	} else {
		d.Constraint = r
	}

	return d
}

// UnmarshalYAML is a hook for gopkg.in/yaml.v2 in the unmarshaling process
func (d *Dependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	newDep := dep{}
	err := unmarshal(&newDep)
	if err != nil {
		return err
	}

	d.Name = newDep.Name
	d.Repository = newDep.Repository
	d.Arch = newDep.Arch
	d.Os = newDep.Os

	if newDep.Reference != "" {
		r := newDep.Reference
		// TODO(sdboyer) this covers git & hg; bzr and svn (??) need love
		if len(r) == 40 {
			if _, err := hex.DecodeString(r); err == nil {
				d.Constraint = gps.Revision(r)
			}
		} else {
			d.Constraint, err = gps.NewSemverConstraint(r)
			if err != nil {
				d.Constraint = gps.NewVersion(r)
			}
		}

		if err != nil {
			return fmt.Errorf("Error on creating constraint for %q from %q: %s", d.Name, r, err)
		}
	} else if newDep.Branch != "" {
		d.Constraint = gps.NewBranch(newDep.Branch)

		if err != nil {
			return fmt.Errorf("Error on creating constraint for %q from %q: %s", d.Name, newDep.Branch, err)
		}
	} else {
		d.Constraint = gps.Any()
	}

	return nil
}

// MarshalYAML is a hook for gopkg.in/yaml.v2 in the marshaling process
func (d *Dependency) MarshalYAML() (interface{}, error) {
	newDep := &dep{
		Name:       d.Name,
		Repository: d.Repository,
		Arch:       d.Arch,
		Os:         d.Os,
	}

	// Pull out the correct type of constraint
	if v, ok := d.Constraint.(gps.Version); ok {
		switch v.Type() {
		case "any":
			// Do nothing; nothing here is taken as 'any'
		case "branch":
			newDep.Branch = v.String()
		case "revision", "semver", "version":
			newDep.Reference = v.String()
		}
	} else if gps.IsAny(d.Constraint) {
		// We do nothing here, as the way any gets represented is with no
		// constraint information at all
	} else if d.Constraint != nil {
		// The only other thing this could really be is a semver range. This
		// will dump that appropriately.
		newDep.Reference = d.Constraint.String()
	}
	// Just ignore any other case

	return newDep, nil
}

// GetRepo retrieves a Masterminds/vcs repo object configured for the root
// of the package being retrieved.
// TODO remove
func (d *Dependency) GetRepo(dest string) (vcs.Repo, error) {
	// The remote location is either the configured repo or the package
	// name as an https url.
	var remote string
	if len(d.Repository) > 0 {
		remote = d.Repository
	} else {
		remote = "https://" + d.Name
	}

	// If the VCS type has a value we try that first.
	if len(d.VcsType) > 0 && d.VcsType != "None" {
		switch vcs.Type(d.VcsType) {
		case vcs.Git:
			return vcs.NewGitRepo(remote, dest)
		case vcs.Svn:
			return vcs.NewSvnRepo(remote, dest)
		case vcs.Hg:
			return vcs.NewHgRepo(remote, dest)
		case vcs.Bzr:
			return vcs.NewBzrRepo(remote, dest)
		default:
			return nil, fmt.Errorf("Unknown VCS type %s set for %s", d.VcsType, d.Name)
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

// HasSubpackage returns if the subpackage is present on the dependency
// TODO remove
func (d *Dependency) HasSubpackage(sub string) bool {
	return false
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
