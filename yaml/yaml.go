// Package yaml provides the ability to work with glide.yaml files.
package yaml

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/glide/util"
	"github.com/Masterminds/vcs"
	"gopkg.in/yaml.v2"
)

// FromYaml takes a yaml string and converts it to a Config instance.
func FromYaml(yml string) (*Config, error) {
	c := &Config{}
	err := yaml.Unmarshal([]byte(yml), &c)
	if err != nil {
		return nil, err
	}

	// The ref property is for the legacy yaml file structure.
	// This sets the currect version to the ref if a version isn't
	// already set.
	for _, v := range c.Imports {
		if v.Reference == "" && v.Ref != "" {
			v.Reference = v.Ref
		}
		v.Ref = ""

		// Make sure only legitimate VCS are listed.
		v.VcsType = filterVcsType(v.VcsType)

		// Get the root name for the package
		o := v.Name
		v.Name = util.GetRootFromPackage(v.Name)
		subpkg := strings.TrimPrefix(o, v.Name)
		if len(subpkg) > 0 && subpkg != o {
			v.Subpackages = append(v.Subpackages, strings.TrimPrefix(subpkg, "/"))
		}
	}
	for _, v := range c.DevImports {
		if v.Reference == "" && v.Ref != "" {
			v.Reference = v.Ref
		}
		v.Ref = ""

		v.VcsType = filterVcsType(v.VcsType)

		// Get the root name for the package
		o := v.Name
		v.Name = util.GetRootFromPackage(v.Name)
		subpkg := strings.TrimPrefix(o, v.Name)
		if len(subpkg) > 0 && subpkg != o {
			v.Subpackages = append(v.Subpackages, subpkg)
		}
	}

	c.Imports, err = c.Imports.DeDupe()
	if err != nil {
		return c, err
	}
	c.DevImports, err = c.DevImports.DeDupe()
	if err != nil {
		return c, err
	}

	return c, nil
}

// ToYaml takes a *Config instance and converts it into a yaml string.
func ToYaml(cfg *Config) (string, error) {
	yml, err := yaml.Marshal(&cfg)
	if err != nil {
		return "", err
	}
	return string(yml), nil
}

// Marhsal encodes using the package's YAML encoder.
//
// Use this instead of ToYaml.
func Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

// Config is the top-level configuration object.
type Config struct {
	Parent     *Config      `yaml:"-"`
	Name       string       `yaml:"package"`
	Imports    Dependencies `yaml:"import"`
	DevImports Dependencies `yaml:"devimport,omitempty"`
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

// HasRecursiveDependency returns true if this config or one of it's parents has this dependency
func (c *Config) HasRecursiveDependency(name string) bool {
	if c.HasDependency(name) == true {
		return true
	} else if c.Parent != nil {
		return c.Parent.HasRecursiveDependency(name)
	}
	return false
}

// GetRoot follows the Parent down to the top node
func (c *Config) GetRoot() *Config {
	if c.Parent != nil {
		return c.Parent.GetRoot()
	}
	return c
}

// Clone performs a deep clone of the Config instance
func (c *Config) Clone() *Config {
	n := &Config{}
	n.Name = c.Name
	n.Imports = c.Imports.Clone()
	n.DevImports = c.DevImports.Clone()
	return n
}

// Dependencies is a collection of Dependency
type Dependencies []*Dependency

// Dependency describes a package that the present package depends upon.
type Dependency struct {
	Name             string   `yaml:"package"`
	Reference        string   `yaml:"version,omitempty"`
	Ref              string   `yaml:"ref,omitempty"`
	Pin              string   `yaml:"pin,omitempty"`
	Repository       string   `yaml:"repo,omitempty"`
	VcsType          string   `yaml:"vcs,omitempty"`
	Subpackages      []string `yaml:"subpackages,omitempty"`
	Arch             []string `yaml:"arch,omitempty"`
	Os               []string `yaml:"os,omitempty"`
	UpdateAsVendored bool     `yaml:"-"`
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

// Get a dependency by name
func (d Dependencies) Get(name string) *Dependency {
	for _, dep := range d {
		if dep.Name == name {
			return dep
		}
	}
	return nil
}

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
