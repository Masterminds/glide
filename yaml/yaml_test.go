package yaml

import "testing"

var yml = `
package: fake/testing
import:
  - package: github.com/kylelemons/go-gypsy
    subpackages:
      - yaml
  # Intentionally left spaces at end of next line.
  - package: github.com/Masterminds/convert
    repo: git@github.com:Masterminds/convert.git
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

devimport:
  - package: github.com/kylelemons/go-gypsy
`

func TestFromYaml(t *testing.T) {
	cfg, err := FromYaml(yml)
	if err != nil {
		t.Errorf("Unexpected error parsing yaml %s", err)
	}

	if cfg.Name != "fake/testing" {
		t.Errorf("Inaccurate name found %s", cfg.Name)
	}

	found := false
	for _, i := range cfg.Imports {
		if i.Name == "github.com/Masterminds/cookoo" {
			found = true
		}
	}
	if !found {
		t.Error("Unable to find github.com/Masterminds/cookoo")
	}
}

func TestToYaml(t *testing.T) {
	cfg, err := FromYaml(yml)
	if err != nil {
		t.Errorf("Unexpected error parsing yaml %s", err)
	}

	o, err := ToYaml(cfg)
	if err != nil {
		t.Errorf("Unexpected error converting cfg to yaml %s", err)
	}

	if o == "" {
		t.Error("Yaml output not generated when expected")
	}
}
