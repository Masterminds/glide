package cfg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/glide/util"
	"github.com/Masterminds/vcs"
	"gopkg.in/yaml.v2"
)

// Config is the top-level configuration object.
type Config struct {
	Name       string       `yaml:"package"`
	Imports    Dependencies `yaml:"import"`
	DevImports Dependencies `yaml:"devimport,omitempty"`
}

// A transitive representation of a dependency for importing and exploting to yaml.
type cf struct {
	Name       *string       `yaml:"package"`
	Imports    *Dependencies `yaml:"import"`
	DevImports *Dependencies `yaml:"devimport,omitempty"`
}

// ConfigFromYaml returns an instance of Config from YAML
func ConfigFromYaml(yml []byte) (*Config, error) {
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(yml), &cfg)
	return cfg, err
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
	newConfig := &cf{
		&c.Name,
		&c.Imports,
		&c.DevImports,
	}
	if err := unmarshal(&newConfig); err != nil {
		return err
	}

	// Cleanup the Config object now that we have it.
	err := c.DeDupe()

	return err
}

// MarshalYAML is a hook for gopkg.in/yaml.v2 in the marshaling process
func (c *Config) MarshalYAML() (interface{}, error) {
	newConfig := &cf{
		Name: &c.Name,
	}
	i, err := c.Imports.Clone().DeDupe()
	if err != nil {
		return newConfig, err
	}

	di, err := c.DevImports.Clone().DeDupe()
	if err != nil {
		return newConfig, err
	}

	newConfig.Imports = &i
	newConfig.DevImports = &di

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

// Clone performs a deep clone of the Config instance
func (c *Config) Clone() *Config {
	n := &Config{}
	n.Name = c.Name
	n.Imports = c.Imports.Clone()
	n.DevImports = c.DevImports.Clone()
	return n
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
		c.Imports = append(c.DevImports[:found], c.DevImports[found+1:]...)
	}

	return nil
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

// Clone performs a deep clone of Dependencies
func (d Dependencies) Clone() Dependencies {
	n := make(Dependencies, 0, 1)
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
			if dep.Reference != v.Reference {
				return d, fmt.Errorf("Import %s repeated with different versions '%s' and '%s'", dep.Name, dep.Reference, v.Reference)
			}
			if dep.Repository != v.Repository || dep.VcsType != v.VcsType {
				return d, fmt.Errorf("Import %s repeated with different Repository details", dep.Name)
			}
			if !reflect.DeepEqual(dep.Os, v.Os) || !reflect.DeepEqual(dep.Arch, v.Arch) {
				return d, fmt.Errorf("Import %s repeated with different OS or Architecture filtering", dep.Name)
			}
			imports[checked[dep.Name]].Subpackages = stringArrayDeDupe(v.Subpackages, dep.Subpackages...)
		}
	}

	return imports, nil
}

// Dependency describes a package that the present package depends upon.
type Dependency struct {
	Name             string   `yaml:"package"`
	Reference        string   `yaml:"version,omitempty"`
	Pin              string   `yaml:"-"`
	Repository       string   `yaml:"repo,omitempty"`
	VcsType          string   `yaml:"vcs,omitempty"`
	Subpackages      []string `yaml:"subpackages,omitempty"`
	Arch             []string `yaml:"arch,omitempty"`
	Os               []string `yaml:"os,omitempty"`
	UpdateAsVendored bool     `yaml:"-"`
}

// A transitive representation of a dependency for importing and exploting to yaml.
type dep struct {
	Name        *string   `yaml:"package"`
	Reference   *string   `yaml:"version,omitempty"`
	Ref         *string   `yaml:"ref,omitempty"`
	Repository  *string   `yaml:"repo,omitempty"`
	VcsType     *string   `yaml:"vcs,omitempty"`
	Subpackages *[]string `yaml:"subpackages,omitempty"`
	Arch        *[]string `yaml:"arch,omitempty"`
	Os          *[]string `yaml:"os,omitempty"`
}

// UnmarshalYAML is a hook for gopkg.in/yaml.v2 in the unmarshaling process
func (d *Dependency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ref string
	newDep := &dep{
		&d.Name,
		&d.Reference,
		&ref,
		&d.Repository,
		&d.VcsType,
		&d.Subpackages,
		&d.Arch,
		&d.Os,
	}
	err := unmarshal(&newDep)
	if err != nil {
		return err
	}

	if d.Reference == "" && ref != "" {
		d.Reference = ref
	}

	// Make sure only legitimate VCS are listed.
	d.VcsType = filterVcsType(d.VcsType)

	// Get the root name for the package
	o := d.Name
	d.Name = util.GetRootFromPackage(d.Name)
	subpkg := strings.TrimPrefix(o, d.Name)
	if len(subpkg) > 0 && subpkg != o {
		d.Subpackages = append(d.Subpackages, strings.TrimPrefix(subpkg, "/"))
	}

	return nil
}

// MarshalYAML is a hook for gopkg.in/yaml.v2 in the marshaling process
func (d *Dependency) MarshalYAML() (interface{}, error) {

	// Make sure we only write the correct vcs type to file
	t := filterVcsType(d.VcsType)
	newDep := &dep{
		Name:        &d.Name,
		Reference:   &d.Reference,
		Repository:  &d.Repository,
		VcsType:     &t,
		Subpackages: &d.Subpackages,
		Arch:        &d.Arch,
		Os:          &d.Os,
	}

	return newDep, nil
}

// GetRepo retrieves a Masterminds/vcs repo object configured for the root
// of the package being retrieved.
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
	return &Dependency{
		Name:             d.Name,
		Reference:        d.Reference,
		Pin:              d.Pin,
		Repository:       d.Repository,
		VcsType:          d.VcsType,
		Subpackages:      d.Subpackages,
		Arch:             d.Arch,
		Os:               d.Os,
		UpdateAsVendored: d.UpdateAsVendored,
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
