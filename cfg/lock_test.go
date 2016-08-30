package cfg

import (
	"testing"

	"github.com/kr/pretty"
)

const lockFix = `
imports:
- name: github.com/gogo/protobuf
  revision: 82d16f734d6d871204a3feb1a73cb220cc92574c
`

const llockFix = `
imports:
- name: github.com/gogo/protobuf
  version: 82d16f734d6d871204a3feb1a73cb220cc92574c
`

func TestLegacyLockAutoconvert(t *testing.T) {
	ll, legacy, err := LockfileFromYaml([]byte(llockFix))
	if err != nil {
		t.Errorf("LockfileFromYaml failed to detect and autoconvert legacy lock file with err %s", err)
	}
	pretty.Println(ll)
	if !legacy {
		t.Error("LockfileFromYaml failed to report autoconversion of legacy lock file")
	}

	if len(ll.Imports) != 1 {
		t.Errorf("LockfileFromYaml autoconverted with wrong number of import stanzas; expected 1, got %v", len(ll.Imports))
	}
}
