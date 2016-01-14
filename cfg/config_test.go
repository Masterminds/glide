package cfg

import (
	"testing"

	"gopkg.in/yaml.v2"
)

var yml = `
package: fake/testing
description: foo bar baz
homepage: https://example.com
license: MIT
owners:
- name: foo
  email: bar@example.com
  homepage: https://example.com
import:
  - package: github.com/kylelemons/go-gypsy
    subpackages:
      - yaml
  # Intentionally left spaces at end of next line.
  - package: github.com/Masterminds/convert
    repo: git@github.com:Masterminds/convert.git
    vcs: git
    ref: a9949121a2e2192ca92fa6dddfeaaa4a4412d955
    subpackages:
      - color
      - nautical
      - radial
    os:
      - linux
    arch:
      - i386
      - arm
  - package: github.com/Masterminds/structable
  - package: github.com/Masterminds/cookoo/color
  - package: github.com/Masterminds/cookoo/convert

devimport:
  - package: github.com/kylelemons/go-gypsy
`

func TestManualConfigFromYaml(t *testing.T) {
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(yml), &cfg)
	if err != nil {
		t.Errorf("Unable to Unmarshal config yaml")
	}

	if cfg.Name != "fake/testing" {
		t.Errorf("Inaccurate name found %s", cfg.Name)
	}

	if cfg.Description != "foo bar baz" {
		t.Errorf("Inaccurate description found %s", cfg.Description)
	}

	if cfg.Home != "https://example.com" {
		t.Errorf("Inaccurate homepage found %s", cfg.Home)
	}

	if cfg.License != "MIT" {
		t.Errorf("Inaccurate license found %s", cfg.License)
	}

	found := false
	found2 := false
	for _, i := range cfg.Imports {
		if i.Name == "github.com/Masterminds/convert" {
			found = true
			ref := "a9949121a2e2192ca92fa6dddfeaaa4a4412d955"
			if i.Reference != ref {
				t.Errorf("Config reference for cookoo is inaccurate. Expected '%s' found '%s'", ref, i.Reference)
			}
		}

		if i.Name == "github.com/Masterminds/cookoo" {
			found2 = true
			if i.Subpackages[0] != "color" {
				t.Error("Dependency separating package and subpackage not working")
			}
		}
	}
	if !found {
		t.Error("Unable to find github.com/Masterminds/convert")
	}
	if !found2 {
		t.Error("Unable to find github.com/Masterminds/cookoo")
	}
}

func TestClone(t *testing.T) {
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(yml), &cfg)
	if err != nil {
		t.Errorf("Unable to Unmarshal config yaml")
	}

	cfg2 := cfg.Clone()
	if cfg2.Name != "fake/testing" {
		t.Error("Config cloning failed")
	}
	if cfg2.License != "MIT" {
		t.Error("Config cloning failed to copy License")
	}
	cfg.Name = "foo"

	if cfg.Name == cfg2.Name {
		t.Error("Cloning Config name failed")
	}
}

func TestConfigFromYaml(t *testing.T) {
	c, err := ConfigFromYaml([]byte(yml))
	if err != nil {
		t.Error("ConfigFromYaml failed to parse yaml")
	}

	if c.Name != "fake/testing" {
		t.Error("ConfigFromYaml failed to properly parse yaml")
	}
}

func TestHasDependency(t *testing.T) {
	c, err := ConfigFromYaml([]byte(yml))
	if err != nil {
		t.Error("ConfigFromYaml failed to parse yaml for HasDependency")
	}

	if c.HasDependency("github.com/Masterminds/convert") != true {
		t.Error("HasDependency failing to pickup depenency")
	}

	if c.HasDependency("foo/bar/bar") != false {
		t.Error("HasDependency picking up dependency it shouldn't")
	}
}

func TestOwners(t *testing.T) {
	o := new(Owner)
	o.Name = "foo"
	o.Email = "foo@example.com"
	o.Home = "https://foo.example.com"

	o2 := o.Clone()
	if o2.Name != o.Name || o2.Email != o.Email || o2.Home != o.Home {
		t.Error("Unable to clone Owner")
	}

	o.Name = "Bar"
	if o.Name == o2.Name {
		t.Error("Owner clone is a pointer instead of a clone")
	}

	s := make(Owners, 0, 1)
	s = append(s, o)
	s2 := s.Clone()
	o3 := s2[0]

	o3.Name = "Qux"

	if o3.Name == o.Name {
		t.Error("Owners cloning isn't deep")
	}

	cfg := &Config{}
	err := yaml.Unmarshal([]byte(yml), &cfg)
	if err != nil {
		t.Errorf("Unable to Unmarshal config yaml")
	}

	if cfg.Owners[0].Name != "foo" ||
		cfg.Owners[0].Email != "bar@example.com" ||
		cfg.Owners[0].Home != "https://example.com" {
		t.Error("Unable to parse owners from yaml")
	}
}
