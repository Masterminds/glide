package cmd

import (
	"testing"

	"github.com/Masterminds/cookoo"
)

var yamlFile = `
package: fake/testing
import:
  - package: github.com/kylelemons/go-gypsy
    subpackages: yaml
    flatten: true
  # Intentionally left spaces at end of next line.
  - package: github.com/Masterminds/convert
    repo: git@github.com:Masterminds/convert.git
    ref: a9949121a2e2192ca92fa6dddfeaaa4a4412d955
    subpackages:
      - color
      - nautical
      - radial
    os: linux
    arch:
      - i386
      - arm
    flatten: true
  - package: github.com/Masterminds/structable

devimport:
  - package: github.com/kylelemons/go-gypsy
`

var childYamlFile = `
package: fake/testing/more
import:
  - package: github.com/kylelemons/go-gypsy
    subpackages: yaml
`

func TestFromYaml(t *testing.T) {
	reg, router, cxt := cookoo.Cookoo()

	reg.Route("t", "Testing").
		Does(ParseYamlString, "cfg").Using("yaml").WithDefault(yamlFile).
		Does(ParseYamlString, "childCfg").Using("yaml").WithDefault(childYamlFile)

	if err := router.HandleRequest("t", cxt, false); err != nil {
		t.Errorf("Failed to parse YAML: %s", err)
	}

	cfg := cxt.Get("cfg", nil).(*Config)
	cfgChild := cxt.Get("childCfg", nil).(*Config)
	cfgChild.Parent = cfg

	if cfg.Name != "fake/testing" {
		t.Errorf("Expected name to be 'fake/teting', not '%s'", cfg.Name)
	}

	if len(cfg.Imports) != 3 {
		t.Errorf("Expected 3 imports, got %d", len(cfg.Imports))
	}

	if cfg.Parent != nil {
		t.Error("Expected root glide Parent to be nil")
	}

	if cfg.Imports.Get("github.com/Masterminds/convert") == nil {
		t.Error("Expected Imports.Get to return Dependency")
	}

	if cfg.Imports.Get("github.com/doesnot/exist") != nil {
		t.Error("Execpted Imports.Get to return nil")
	}

	if cfgChild.HasRecursiveDependency("github.com/Masterminds/convert") == false {
		t.Errorf("Expected to find a recursive dependency")
	}

	var imp *Dependency
	for _, d := range cfg.Imports {
		if d.Name == "github.com/Masterminds/convert" {
			imp = d
		}
	}

	if imp == nil {
		t.Errorf("Expected the convert package, got nothing")
	}

	if len(imp.Subpackages) != 3 {
		t.Errorf("Expected 3 subpackages. got %d", len(imp.Subpackages))
	}

	if imp.Subpackages[0] != "color" {
		t.Errorf("Expected first subpackage to be 'color', got '%s'", imp.Subpackages[0])
	}

	if len(imp.Os) != 1 {
		t.Errorf("Expected Os: SOMETHING")
	} else if imp.Os[0] != "linux" {
		t.Errorf("Expected Os: linux")
	}

	if len(imp.Arch) != 2 {
		t.Error("Expected two Archs.")
	} else if imp.Arch[0] != "i386" {
		t.Errorf("Expected arch 1 to be i386, got %s.", imp.Arch[0])
	} else if imp.Arch[1] != "arm" {
		t.Error("Expected arch 2 to be arm.")
	}

	if imp.Repository != "git@github.com:Masterminds/convert.git" {
		t.Errorf("Got wrong repo")
	}
	if imp.Reference != "a9949121a2e2192ca92fa6dddfeaaa4a4412d955" {
		t.Errorf("Got wrong reference.")
	}

	if imp.Flatten != true {
		t.Errorf("Expected Flatten: true")
	}

	if len(cfg.DevImports) != 1 {
		t.Errorf("Expected one dev import.")
	}

}

func TestNormalizeName(t *testing.T) {
	packages := map[string]string{
		"github.com/Masterminds/cookoo/web/io/foo": "github.com/Masterminds/cookoo",
		"golang.org/x/crypto/ssh":                  "golang.org/x/crypto",
		"technosophos.me/x/totally/fake/package":   "technosophos.me/x/totally",
		"incomplete/example":                       "incomplete/example",
		"net":                                      "net",
	}
	for start, expected := range packages {
		if finish, extra := NormalizeName(start); expected != finish {
			t.Errorf("Expected '%s', got '%s'", expected, finish)
		} else if start != finish && start != finish+"/"+extra {
			t.Errorf("Expected %s to end with %s", finish, extra)
		}
	}
}
