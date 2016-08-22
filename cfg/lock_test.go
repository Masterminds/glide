package cfg

import (
	"sort"
	"strings"
	"testing"
)

func TestSortLocks(t *testing.T) {
	c, err := ConfigFromYaml([]byte(yml))
	if err != nil {
		t.Error("ConfigFromYaml failed to parse yaml for TestSortDependencies")
	}

	ls := make(Locks, len(c.Imports))
	for i := 0; i < len(c.Imports); i++ {
		ls[i] = &Lock{
			Name:    c.Imports[i].Name,
			Version: c.Imports[i].Reference,
		}
	}

	if ls[2].Name != "github.com/Masterminds/structable" {
		t.Error("Initial dependencies are out of order prior to sort")
	}

	sort.Sort(ls)

	if ls[0].Name != "github.com/kylelemons/go-gypsy" ||
		ls[1].Name != "github.com/Masterminds/convert" ||
		ls[2].Name != "github.com/Masterminds/cookoo" ||
		ls[3].Name != "github.com/Masterminds/structable" {
		t.Error("Sorting of dependencies failed")
	}
}

const inputSubpkgYaml = `
imports:
- name: github.com/gogo/protobuf
  version: 82d16f734d6d871204a3feb1a73cb220cc92574c
  subpackages:
  - plugin/equal
  - sortkeys
  - plugin/face
  - plugin/gostring
  - vanity
  - plugin/grpc
  - plugin/marshalto
  - plugin/populate
  - plugin/oneofcheck
  - plugin/size
  - plugin/stringer
  - plugin/defaultcheck
  - plugin/embedcheck
  - plugin/description
  - plugin/enumstringer
  - gogoproto
  - plugin/testgen
  - plugin/union
  - plugin/unmarshal
  - protoc-gen-gogo/generator
  - protoc-gen-gogo/plugin
  - vanity/command
  - protoc-gen-gogo/descriptor
  - proto
`
const expectSubpkgYaml = `
imports:
- name: github.com/gogo/protobuf
  version: 82d16f734d6d871204a3feb1a73cb220cc92574c
  subpackages:
  - gogoproto
  - plugin/defaultcheck
  - plugin/description
  - plugin/embedcheck
  - plugin/enumstringer
  - plugin/equal
  - plugin/face
  - plugin/gostring
  - plugin/grpc
  - plugin/marshalto
  - plugin/oneofcheck
  - plugin/populate
  - plugin/size
  - plugin/stringer
  - plugin/testgen
  - plugin/union
  - plugin/unmarshal
  - proto
  - protoc-gen-gogo/descriptor
  - protoc-gen-gogo/generator
  - protoc-gen-gogo/plugin
  - sortkeys
  - vanity
  - vanity/command
`

func TestSortSubpackages(t *testing.T) {
	lf, err := LockfileFromYaml([]byte(inputSubpkgYaml))
	if err != nil {
		t.Fatal(err)
	}

	out, err := lf.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(out), expectSubpkgYaml) {
		t.Errorf("Expected %q\n to contain\n%q", string(out), expectSubpkgYaml)
	}
}
