package cmd

import (
	"github.com/Masterminds/cookoo"
	"github.com/kylelemons/go-gypsy/yaml"
	"fmt"
)

func ParseYaml(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	conf := new(Config)
	f, err := yaml.ReadFile("./glide.yaml")
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
	}

	conf.Imports = make([]*Dependency, 0, 1)
	if imp, ok := vals["import"]; ok {
		imports := imp.(yaml.List)
		for _, v := range imports {
			pkg := v.(yaml.Map)
			dep := Dependency {
				Name: valOrEmpty("package", pkg),
				Reference: valOrEmpty("ref", pkg),
				VcsType: valOrEmpty("vcs", pkg),
				Repository: valOrEmpty("repo", pkg),
			}
			conf.Imports = append(conf.Imports, &dep)
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

// Config is the top-level configuration object.
type Config struct {
	Name string
	Imports []*Dependency
	DevImports []*Dependency
}

// Dependency describes a package that the present package depends upon.
type Dependency struct {
	Name, Reference, Repository, VcsType string
}
