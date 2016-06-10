package dependency

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Masterminds/glide/cfg"
)

func TestResolveLocalShallow(t *testing.T) {
	r, err := NewResolver("../")
	if err != nil {
		t.Fatal(err)
	}

	l, _, err := r.ResolveLocal(false)
	if err != nil {
		t.Fatalf("Failed to resolve: %s", err)
	}

	expect := []string{
		"github.com/Masterminds/semver",
		"github.com/Masterminds/vcs",
		"gopkg.in/yaml.v2",
		"github.com/codegangsta/cli",
	}

	for _, p := range expect {
		found := false
		for _, li := range l {
			if strings.HasSuffix(li, p) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Could not find %s in resolved list.", p)
		}
	}
}

func TestResolveLocalDeep(t *testing.T) {
	r, err := NewResolver("../")
	if err != nil {
		t.Fatal(err)
	}

	l, _, err := r.ResolveLocal(true)
	if err != nil {
		t.Fatalf("Failed to resolve: %s", err)
	}

	if len(l) < 4 {
		t.Errorf("Expected at least 4 deps, got %d", len(l))
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

	if !strings.HasSuffix("github.com/codegangsta/cli", l[0]) {
		t.Errorf("Unexpected package name: %s", l[0])
	}
}

func TestResolveAll(t *testing.T) {
	// These are build dependencies of Glide, so we know they are here.
	deps := []*cfg.Dependency{
		{Name: "github.com/codegangsta/cli"},
		{Name: "github.com/Masterminds/semver"},
		{Name: "github.com/Masterminds/vcs"},
		{Name: "gopkg.in/yaml.v2"},
	}

	r, err := NewResolver("../")
	if err != nil {
		t.Fatalf("No new resolver: %s", err)
	}
	l, err := r.ResolveAll(deps, false)
	if err != nil {
		t.Fatalf("Failed to resolve: %s", err)
	}

	if len(l) < len(deps) {
		t.Errorf("Expected at least %d deps, got %d", len(deps), len(l))
	}
}
