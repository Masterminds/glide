package cfg

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/glide/util"
)

// lConfig1 is a legacy Config file.
type lConfig1 struct {
	Name        string         `yaml:"package"`
	Description string         `json:"description,omitempty"`
	Home        string         `yaml:"homepage,omitempty"`
	License     string         `yaml:"license,omitempty"`
	Owners      Owners         `yaml:"owners,omitempty"`
	Ignore      []string       `yaml:"ignore,omitempty"`
	Exclude     []string       `yaml:"excludeDirs,omitempty"`
	Imports     lDependencies1 `yaml:"import"`
	DevImports  lDependencies1 `yaml:"testImport,omitempty"`
}

func (c *lConfig1) UnmarshalYAML(unmarshal func(interface{}) error) error {
	newConfig := &lcf1{}
	if err := unmarshal(&newConfig); err != nil {
		return err
	}
	c.Name = newConfig.Name
	c.Description = newConfig.Description
	c.Home = newConfig.Home
	c.License = newConfig.License
	c.Owners = newConfig.Owners
	c.Ignore = newConfig.Ignore
	c.Exclude = newConfig.Exclude
	c.Imports = newConfig.Imports
	c.DevImports = newConfig.DevImports

	// Cleanup the Config object now that we have it.
	err := c.DeDupe()

	return err
}

// DeDupe consolidates duplicate dependencies on a Config instance
func (c *lConfig1) DeDupe() error {

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

// Legacy representation of a glide.yaml file.
type lcf1 struct {
	Name        string         `yaml:"package"`
	Description string         `yaml:"description,omitempty"`
	Home        string         `yaml:"homepage,omitempty"`
	License     string         `yaml:"license,omitempty"`
	Owners      Owners         `yaml:"owners,omitempty"`
	Ignore      []string       `yaml:"ignore,omitempty"`
	Exclude     []string       `yaml:"excludeDirs,omitempty"`
	Imports     lDependencies1 `yaml:"import"`
	DevImports  lDependencies1 `yaml:"testImport,omitempty"`
}

type lDependencies1 []*lDependency1

type lDependency1 struct {
	Name        string   `yaml:"package"`
	Reference   string   `yaml:"version,omitempty"`
	Pin         string   `yaml:"-"`
	Repository  string   `yaml:"repo,omitempty"`
	VcsType     string   `yaml:"vcs,omitempty"`
	Subpackages []string `yaml:"subpackages,omitempty"`
	Arch        []string `yaml:"arch,omitempty"`
	Os          []string `yaml:"os,omitempty"`
}

// Legacy unmarshaler
func (d *lDependency1) UnmarshalYAML(unmarshal func(interface{}) error) error {
	newDep := &ldep1{}
	err := unmarshal(&newDep)
	if err != nil {
		return err
	}
	d.Name = newDep.Name
	d.Reference = newDep.Reference
	d.Repository = newDep.Repository
	d.VcsType = newDep.VcsType
	d.Subpackages = newDep.Subpackages
	d.Arch = newDep.Arch
	d.Os = newDep.Os

	if d.Reference == "" && newDep.Ref != "" {
		d.Reference = newDep.Ref
	}

	// Make sure only legitimate VCS are listed.
	d.VcsType = filterVcsType(d.VcsType)

	// Get the root name for the package
	tn, subpkg := util.NormalizeName(d.Name)
	d.Name = tn
	if subpkg != "" {
		d.Subpackages = append(d.Subpackages, subpkg)
	}

	// Older versions of Glide had a / prefix on subpackages in some cases.
	// Here that's cleaned up. Someday we should be able to remove this.
	for k, v := range d.Subpackages {
		d.Subpackages[k] = strings.TrimPrefix(v, "/")
	}

	return nil
}

// DeDupe cleans up duplicates on a list of dependencies.
func (d lDependencies1) DeDupe() (lDependencies1, error) {
	checked := map[string]int{}
	imports := make(lDependencies1, 0, 1)
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

// Legacy representation of a dep constraint
type ldep1 struct {
	Name        string   `yaml:"package"`
	Reference   string   `yaml:"version,omitempty"`
	Ref         string   `yaml:"ref,omitempty"`
	Repository  string   `yaml:"repo,omitempty"`
	VcsType     string   `yaml:"vcs,omitempty"`
	Subpackages []string `yaml:"subpackages,omitempty"`
	Arch        []string `yaml:"arch,omitempty"`
	Os          []string `yaml:"os,omitempty"`
}

type lLock1 struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Repository  string   `yaml:"repo,omitempty"`
	VcsType     string   `yaml:"vcs,omitempty"`
	Subpackages []string `yaml:"subpackages,omitempty"`
	Arch        []string `yaml:"arch,omitempty"`
	Os          []string `yaml:"os,omitempty"`
}
