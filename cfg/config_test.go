package cfg

import (
	"testing"

	"gopkg.in/yaml.v2"
)

var yml = `
package: fake/testing
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

func TestConfigFromYaml(t *testing.T) {
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(yml), &cfg)
	if err != nil {
		t.Errorf("Unable to Unmarshal config yaml")
	}

	if cfg.Name != "fake/testing" {
		t.Errorf("Inaccurate name found %s", cfg.Name)
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
	cfg.Name = "foo"

	if cfg.Name == cfg2.Name {
		t.Error("Cloning Config name failed")
	}
}
