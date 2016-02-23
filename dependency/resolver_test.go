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

	l, err := r.ResolveLocal(false)
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

	l, err := r.ResolveLocal(true)
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
	l, err := r.ResolveAll(deps)
	if err != nil {
		t.Fatalf("Failed to resolve: %s", err)
	}

	if len(l) < len(deps) {
		t.Errorf("Expected at least %d deps, got %d", len(deps), len(l))
	}
}

func TestSliceToQueueNoSubpackages(t *testing.T) {
	basepath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/Masterminds/glide/vendor")
	pkg := "github.com/codegangsta/cli"
	fullpath := filepath.Join(basepath, pkg)
	deps := []*cfg.Dependency{
		{Name: pkg},
	}
	l := sliceToQueue(deps, basepath)
	if l.Len() != len(deps) {
		t.Fatalf("Wrong number of queue items: want %d, got %d", len(deps), l.Len())
	}
	if s := l.Front().Value.(string); s != fullpath {
		t.Errorf("Wrong value as queue head: want %s, got [%s]", fullpath, s)
	}
}

func TestSliceToQueueWithSubpackages(t *testing.T) {
	basepath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/Masterminds/glide/vendor")
	pkg := "golang.org/x/crypto"
	subpkg := "ssh"
	fullpath := filepath.Join(basepath, pkg, subpkg)
	deps := []*cfg.Dependency{
		{Name: pkg, Subpackages: []string{subpkg}},
	}
	l := sliceToQueue(deps, basepath)
	if l.Len() != len(deps) {
		t.Fatalf("Wrong number of queue items: want %d, got %d", len(deps), l.Len())
	}
	if s := l.Front().Value.(string); s != fullpath {
		t.Errorf("Wrong value as queue head: want %s, got %s", fullpath, s)
	}
}

func TestSliceToQueueWithMultSubpackages(t *testing.T) {
	// verify that pkg with multiple subpackages gets expanded into multiple queue entries
	basepath := filepath.Join(os.Getenv("GOPATH"), "src/github.com/Masterminds/glide/vendor")
	pkg := "golang.org/x/crypto"
	subpkgs := []string{"bcrypt", "ssh"}
	fullpath := []string{filepath.Join(basepath, pkg, subpkgs[0]),
		filepath.Join(basepath, pkg, subpkgs[1])}
	deps := []*cfg.Dependency{
		{Name: pkg, Subpackages: subpkgs},
	}
	l := sliceToQueue(deps, basepath)
	if l.Len() != len(deps)+1 {
		t.Fatalf("Wrong number of queue items: want %d, got %d", len(deps)+1, l.Len())
	}
	var cnt int
	for e := l.Front(); e != nil; e = e.Next() {
		if s := e.Value.(string); s != fullpath[cnt] {
			t.Errorf("Wrong value as queue element #%d: want %s, got %s", cnt, fullpath[cnt], s)
		}
		cnt++
	}
}
