package dependency

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/glide/yaml"
)

func TestResolveLocal(t *testing.T) {
	r, err := NewResolver("../")
	if err != nil {
		t.Fatal(err)
	}

	l, err := r.ResolveLocal()
	if err != nil {
		t.Fatalf("Failed to resolve: %s", err)
	}

	for _, p := range l {
		t.Log(p)
	}

	if len(l) != 12 {
		t.Errorf("Expected 12 dep, got %d: %s", len(l))
	}
}

func TestResolve(t *testing.T) {
	r, err := NewResolver("../")
	if err != nil {
		t.Fatal(err)
	}

	base := filepath.Join(os.Getenv("GOPATH"), "src/github.com/Masterminds/glide/vendor")
	l, err := r.Resolve("github.com/codegangsta/cli", base)
	if err != nil {
		t.Fatalf("Failed to resolve: %s", err)
	}

	if len(l) != 1 {
		t.Errorf("Expected 1 dep, got %d: %s", len(l), l[0])
	}
}

func TestResolveAll(t *testing.T) {
	t.Skip()
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
