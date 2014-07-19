package cmd

import (
	"github.com/Masterminds/cookoo"
	"github.com/kylelemons/go-gypsy/yaml"
	"fmt"
)

// ParseYaml parses the glide.yaml format and returns a Configuration object.
func ParseYaml(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fname := p.Get("filename", "glide.yaml").(string)
	conf := new(Config)
	f, err := yaml.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	// Convenience:
	top, ok := f.Root.(yaml.Map)
	if !ok {
		return nil, fmt.Errorf("Expected YAML root to be map, got %t", f.Root)
	}

	vals := map[string]yaml.Node(top)
	if name, ok := vals["package"]; ok {
		//c.Put("cfg.package", name.(yaml.Scalar).String())
		conf.Name = name.(yaml.Scalar).String()
	} else {
		Warn("The 'package' directive is required in Glide YAML.\n")
	}

	// Allow the user to override the behavior of `glide in`.
	if incmd, ok := vals["incmd"]; ok {
		conf.InCommand = incmd.(yaml.Scalar).String()
		//fmt.Printf("[DEBUG] Custom glide in: %s\n", conf.InCommand)
	}

	conf.Imports = make([]*Dependency, 0, 1)
	if imp, ok := vals["import"]; ok {
		imports, ok := imp.(yaml.List)

		if ok {
			for _, v := range imports {
				pkg := v.(yaml.Map)
				dep := Dependency {
					Name: valOrEmpty("package", pkg),
					Reference: valOrEmpty("ref", pkg),
					VcsType: getVcsType(pkg),
					Repository: valOrEmpty("repo", pkg),
					Subpackages: subpkg("subpackages", pkg),
				}
				conf.Imports = append(conf.Imports, &dep)
			}
		}
	}

	return conf, nil
}

func valOrEmpty(key string, store map[string]yaml.Node) string {
	val, ok := store[key]
	if !ok {
		return ""
	}
	return val.(yaml.Scalar).String()
}

func subpkg(key string, store map[string]yaml.Node) []string {
	val, ok := store[key]

	subpackages := []string{}
	if !ok {
		return subpackages
	}

	pkgs, ok := val.(yaml.List)

	if !ok {

		// Special case: Allow 'subpackages: justOne'
		if one, ok := val.(yaml.Scalar); ok {
			return []string{ one.String() }
		}

		Warn("Expected list of subpackages.\n")
		return subpackages
	}


	for _, pkg := range pkgs {
		subpackages = append(subpackages, pkg.(yaml.Scalar).String())
	}
	return subpackages
}

func getVcsType(store map[string]yaml.Node) uint {

	val, ok := store["vcs"]
	if !ok {
		return NoVCS
	}

	name := val.(yaml.Scalar).String()

	switch name {
	case "git":
		return Git
	case "hg", "mercurial":
		return Hg
	case "bzr", "bazaar":
		return Bzr
	case "svn", "subversion":
		return Svn
	default:
		return NoVCS
	}
}

// Config is the top-level configuration object.
type Config struct {
	Name string
	Imports []*Dependency
	DevImports []*Dependency
	// InCommand is the default shell command run to start a 'glide in'
	// session.
	InCommand string
}

// Dependency describes a package that the present package depends upon.
type Dependency struct {
	Name, Reference, Repository string
	VcsType uint
	Subpackages []string
}
