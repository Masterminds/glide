package mirrors

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

func TestSortMirrors(t *testing.T) {
	ov, err := FromYaml([]byte(oyml))
	if err != nil {
		t.Error("Unable to read mirrors yaml")
	}

	out, err := ov.Marshal()
	if err != nil {
		t.Error("Unable to marshal mirrors yaml")
	}

	if string(out) != ooutyml {
		t.Error("Output mirrors sorting failed")
	}
}
