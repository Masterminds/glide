package dependency

import (
	"testing"

	"github.com/Masterminds/glide/yaml"
)

func TestResolveAll(t *testing.T) {
	// These are build dependencies of Glide, so we know they are here.
	deps := []*yaml.Dependency{
		&yaml.Dependency{Name: "github.com/codegangsta/cli"},
		&yaml.Dependency{Name: "github.com/Masterminds/cookoo"},
		&yaml.Dependency{Name: "github.com/Masterminds/squirrel"},
		&yaml.Dependency{Name: "gopkg.in/yaml.v2"},
	}

	r, err := NewResolver("../")
	if err != nil {
		t.Fatalf("No new resolver: %s", err)
	}
	l, err := r.ResolveAll(deps)
	if err != nil {
		t.Fatalf("Failed to resolve: %s", err)
	}

	if len(l) < 3 {
		t.Errorf("Expected len=3, got %d", len(l))
	}

	println("SEEN")
	for k := range r.seen {
		println(k)
	}
	println("RESULT")

	for _, v := range l {
		println(v)
	}
}
