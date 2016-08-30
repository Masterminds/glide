package cfg

import (
	"reflect"
	"testing"

	"github.com/sdboyer/gps"

	"gopkg.in/yaml.v2"
)

var lyml = `
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
    version: v1.0.0
  - package: github.com/Masterminds/cookoo/color
  - package: github.com/Masterminds/cookoo/convert

testImport:
  - package: github.com/kylelemons/go-gypsy
`

var yml = `
package: fake/testing
description: foo bar baz
homepage: https://example.com
license: MIT
owners:
- name: foo
  email: bar@example.com
  homepage: https://example.com
dependencies:
  - package: github.com/kylelemons/go-gypsy
    version: v1.0.0
  - package: github.com/Masterminds/convert
    repo: git@github.com:Masterminds/convert.git
    version: a9949121a2e2192ca92fa6dddfeaaa4a4412d955
  - package: github.com/Masterminds/structable
    branch: master
  - package: github.com/Masterminds/cookoo
    repo: git://github.com/Masterminds/cookoo
  - package: github.com/sdboyer/gps
    version: ^v1.0.0

testDependencies:
  - package: github.com/Sirupsen/logrus
    version: ~v1.0.0
`

func TestManualConfigFromYaml(t *testing.T) {
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(yml), &cfg)
	if err != nil {
		t.Errorf("Unable to Unmarshal config yaml")
	}

	found := make(map[string]bool)
	for _, i := range cfg.Imports {
		found[i.Name] = true

		switch i.Name {
		case "github.com/kylelemons/go-gypsy":
			ref := gps.NewVersion("v1.0.0")
			if i.Constraint != ref {
				t.Errorf("(%s) Expected %q for constraint, got %q", i.Name, ref, i.Constraint)
			}

		case "github.com/Masterminds/convert":
			ref := gps.Revision("a9949121a2e2192ca92fa6dddfeaaa4a4412d955")
			if i.Constraint != ref {
				t.Errorf("(%s) Expected %q for constraint, got %q", i.Name, ref, i.Constraint)
			}

			repo := "git@github.com:Masterminds/convert.git"
			if i.Repository != repo {
				t.Errorf("(%s) Expected %q for repository, got %q", i.Name, repo, i.Repository)
			}

		case "github.com/Masterminds/structable":
			ref := gps.NewBranch("master")
			if i.Constraint != ref {
				t.Errorf("(%s) Expected %q for constraint, got %q", i.Name, ref, i.Constraint)
			}

		case "github.com/Masterminds/cookoo":
			repo := "git://github.com/Masterminds/cookoo"
			if i.Repository != repo {
				t.Errorf("(%s) Expected %q for repository, got %q", i.Name, repo, i.Repository)
			}

		case "github.com/sdboyer/gps":
			sv, _ := gps.NewSemverConstraint("^v1.0.0")
			if !reflect.DeepEqual(sv, i.Constraint) {
				t.Errorf("(%s) Expected %q for constraint, got %q", i.Name, sv, i.Constraint)
			}
		}
	}

	names := []string{
		"github.com/Masterminds/convert",
		"github.com/Masterminds/cookoo",
		"github.com/Masterminds/structable",
		"github.com/kylelemons/go-gypsy",
		"github.com/sdboyer/gps",
	}

	for _, n := range names {
		if !found[n] {
			t.Errorf("Could not find config entry for %s", n)

		}
	}

	if len(cfg.DevImports) != 1 {
		t.Errorf("Expected 1 entry in DevImports, got %v", len(cfg.DevImports))
	} else {
		ti := cfg.DevImports[0]
		n := "github.com/Sirupsen/logrus"
		if ti.Name != n {
			t.Errorf("Expected test dependency to be %s, got %s", n, ti.Name)
		}

		sv, _ := gps.NewSemverConstraint("~v1.0.0")
		if !reflect.DeepEqual(sv, ti.Constraint) {
			t.Errorf("(test dep: %s) Expected %q for constraint, got %q", ti.Name, sv, ti.Constraint)
		}
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

}

func TestLegacyManualConfigFromYaml(t *testing.T) {
	cfg := &lConfig1{}
	err := yaml.Unmarshal([]byte(lyml), &cfg)
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

func TestLegacyConfigAutoconvert(t *testing.T) {
	c, leg, err := ConfigFromYaml([]byte(lyml))
	if err != nil {
		t.Errorf("ConfigFromYaml failed to detect and autoconvert legacy yaml file with err %s", err)
	}

	if !leg {
		t.Errorf("ConfigFromYaml failed to report autoconversion of legacy yaml file")
	}

	if c.Name != "fake/testing" {
		t.Error("ConfigFromYaml failed to properly autoconvert legacy yaml file")
	}

	// Two should survive the conversion
	if len(c.Imports) != 2 {
		t.Error("Expected two dep clauses to survive conversion, but got ", len(c.Imports))
	}

	found := false
	found2 := false
	for _, i := range c.Imports {
		if i.Name == "github.com/Masterminds/convert" {
			found = true
			ref := gps.Revision("a9949121a2e2192ca92fa6dddfeaaa4a4412d955")
			if i.Constraint != ref {
				t.Errorf("(%s) Expected %q for constraint, got %q", i.Name, ref, i.Constraint)
			}

			repo := "git@github.com:Masterminds/convert.git"
			if i.Repository != repo {
				t.Errorf("(%s) Expected %q for repository, got %q", i.Name, repo, i.Repository)
			}
		}

		if i.Name == "github.com/Masterminds/structable" {
			found2 = true
			ref := gps.NewVersion("v1.0.0")
			if i.Constraint != ref {
				t.Errorf("(%s) Expected %q for constraint, got %q", i.Name, ref, i.Constraint)
			}
		}
	}

	if !found {
		t.Error("Unable to find github.com/Masterminds/convert")
	}
	if !found2 {
		t.Error("Unable to find github.com/Masterminds/structable")
	}
}

func TestConfigFromYaml(t *testing.T) {
	c, _, err := ConfigFromYaml([]byte(yml))
	if err != nil {
		t.Error("ConfigFromYaml failed to parse yaml")
	}

	if c.Name != "fake/testing" {
		t.Error("ConfigFromYaml failed to properly parse yaml")
	}
}

func TestHasDependency(t *testing.T) {
	c, _, err := ConfigFromYaml([]byte(yml))
	if err != nil {
		t.Error("ConfigFromYaml failed to parse yaml for HasDependency")
	}

	if !c.HasDependency("github.com/Masterminds/convert") {
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

func TestDeduceConstraint(t *testing.T) {
	// First, valid semver
	c := DeduceConstraint("v1.0.0")
	if c.(gps.Version).Type() != "semver" {
		t.Errorf("Got unexpected version type when passing valid semver string: %T %s", c, c)
	}

	// Now, 20 hex-encoded bytes (which should be assumed to be a SHA1 digest)
	revin := "a9949121a2e2192ca92fa6dddfeaaa4a4412d955"
	c = DeduceConstraint(revin)
	if c != gps.Revision(revin) {
		t.Errorf("Got unexpected version type/val when passing hex-encoded SHA1 digest: %T %s", c, c)
	}

	// Now, the weird bzr guid
	bzrguid := "john@smith.org-20051026185030-93c7cad63ee570df"
	c = DeduceConstraint(bzrguid)
	if c != gps.Revision(bzrguid) {
		t.Errorf("Expected revision with valid bzr guid, got: %T %s", c, c)
	}

	// Check fails if the bzr rev is malformed or weirdly formed
	//
	// chopping off a char should make the hex decode check fail
	c = DeduceConstraint(bzrguid[:len(bzrguid)-1])
	if c != gps.NewVersion(bzrguid[:len(bzrguid)-1]) {
		t.Errorf("Expected plain version when bzr guid has truncated tail hex bits: %T %s", c, c)
	}

	// Extra dash in email doesn't mess us up
	bzrguid2 := "john-smith@smith.org-20051026185030-93c7cad63ee570df"
	c = DeduceConstraint(bzrguid2)
	if c != gps.Revision(bzrguid2) {
		t.Errorf("Expected revision when passing bzr guid has extra dash in email, got: %T %s", c, c)
	}

	// Non-numeric char in middle section bites it
	bzrguid3 := "john-smith@smith.org-2005102a6185030-93c7cad63ee570df"
	c = DeduceConstraint(bzrguid3)
	if c != gps.NewVersion(bzrguid3) {
		t.Errorf("Expected plain version when bzr guid has invalid second section, got: %T %s", c, c)
	}
}
