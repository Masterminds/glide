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
  - package: github.com/technosophos/structable
  # Intentionally left spaces at end of next line.
  - package: github.com/Masterminds/convert   
    repo: git@github.com:Masterminds/convert.git
    ref: a9949121a2e2192ca92fa6dddfeaaa4a4412d955
    subpackages:
      - color
      - nautical
      - radial
devimport:
  - package: github.com/kylelemons/go-gypsy
`

func TestFromYaml(t *testing.T) {
	reg, router, cxt := cookoo.Cookoo()

	reg.Route("t", "Testing").
		Does(ParseYamlString, "cfg").Using("yaml").WithDefault(yamlFile)

	if err := router.HandleRequest("t", cxt, false); err != nil {
		t.Errorf("Failed to parse YAML: %s", err)
	}

	cfg := cxt.Get("cfg", nil).(*Config)
	if cfg.Name != "fake/testing" {
		t.Errorf("Expected name to be 'fake/teting', not '%s'", cfg.Name)
	}

	if len(cfg.Imports) != 3 {
		t.Errorf("Expected 3 imports, got %d", len(cfg.Imports))
	}

	imp := cfg.Imports[2]
	if imp.Name != "github.com/Masterminds/convert" {
		t.Errorf("Expected the convert package, got %s", imp.Name)
	}

	if len(imp.Subpackages) != 3 {
		t.Errorf("Expected 3 subpackages. got %d", len(imp.Subpackages))
	}

	if imp.Subpackages[0]!= "color" {
		t.Errorf("Expected first subpackage to be 'color', got '%s'", imp.Subpackages[0])
	}

	if imp.Repository != "git@github.com:Masterminds/convert.git" {
		t.Errorf("Got wrong repo")
	}
	if imp.Reference != "a9949121a2e2192ca92fa6dddfeaaa4a4412d955" {
		t.Errorf("Got wrong reference.")
	}

	if len(cfg.DevImports) != 1 {
		t.Errorf("Expected one dev import.")
	}


}
