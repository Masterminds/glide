package overrides

import "testing"

var oyml = `
repos:
- original: github.com/Masterminds/semver
  repo: file:///path/to/local/repo
  vcs: git
- original: github.com/Masterminds/atest
  repo: github.com/example/atest
`

var ooutyml = `repos:
- original: github.com/Masterminds/atest
  repo: github.com/example/atest
- original: github.com/Masterminds/semver
  repo: file:///path/to/local/repo
  vcs: git
`

func TestSortOverrides(t *testing.T) {
	ov, err := FromYaml([]byte(oyml))
	if err != nil {
		t.Error("Unable to read overrides yaml")
	}

	out, err := ov.Marshal()
	if err != nil {
		t.Error("Unable to marshal overrides yaml")
	}

	if string(out) != ooutyml {
		t.Error("Output overrides sorting failed")
	}
}
