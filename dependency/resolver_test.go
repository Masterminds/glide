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
		filepath.FromSlash("github.com/Masterminds/semver"),
		filepath.FromSlash("github.com/Masterminds/vcs"),
		filepath.FromSlash("gopkg.in/yaml.v2"),
		filepath.FromSlash("github.com/codegangsta/cli"),
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
	// create package of same name with sys package
	err := os.MkdirAll(filepath.Join(os.Getenv("GOPATH"), "src/strings"), os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create package of test case: %s", err)
	}
	// remove package of same name with sys package
	defer func() {
		err = os.Remove(filepath.Join(os.Getenv("GOPATH"), "src/strings"))
		if err != nil {
			t.Fatalf("Failed to remove package of test case: %s", err)
		}
	}()

	r, err := NewResolver("../")
	if err != nil {
		t.Fatal(err)
	}
	h := &DefaultMissingPackageHandler{Missing: []string{}, Gopath: []string{}, Prefix: "../vendor"}
	r.Handler = h

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
	h := &DefaultMissingPackageHandler{Missing: []string{}, Gopath: []string{}, Prefix: "../vendor"}
	r.Handler = h

	base := filepath.Join(os.Getenv("GOPATH"), "src/github.com/Masterminds/glide/vendor")
	l, err := r.Resolve("github.com/codegangsta/cli", base)
	if err != nil {
		t.Fatalf("Failed to resolve: %s", err)
	}

	if len(l) != 1 {
		t.Errorf("Expected 1 dep, got %d: %s", len(l), l[0])
	}

	if !strings.HasSuffix(filepath.FromSlash("github.com/codegangsta/cli"), l[0]) {
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
	h := &DefaultMissingPackageHandler{Missing: []string{}, Gopath: []string{}, Prefix: "../vendor"}
	r.Handler = h
	l, err := r.ResolveAll(deps, false)
	if err != nil {
		t.Fatalf("Failed to resolve: %s", err)
	}

	if len(l) < len(deps) {
		t.Errorf("Expected at least %d deps, got %d", len(deps), len(l))
	}
}
