package action

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

// Brew converts dependencies to Homebrew resources and prints them, using
// the Lockfile.
//
// Params:
//  - basedir (string): Path to root of project to convert to brew resources
func Brew() {
	base := "."
	// Ensure GOPATH
	EnsureGopath()
	EnsureVendorDir()
	conf := EnsureConfig()

	// Lockfile exists
	if !gpath.HasLock(base) {
		msg.Die("Lock file (glide.lock) does not exist.")
	}
	// Load lockfile
	lock, err := cfg.ReadLockFile(filepath.Join(base, gpath.LockFile))
	if err != nil {
		msg.Die("Could not load lockfile.")
	}
	// Verify lockfile hasn't changed
	hash, err := conf.Hash()
	if err != nil {
		msg.Die("Could not load lockfile.")
	} else if hash != lock.Hash {
		msg.Warn("Lock file may be out of date. Hash check of YAML failed. You may need to run 'update'")
	}

	for _, lock := range lock.Imports {
		resource, err := BrewResourceFromLock(lock)
		if err != nil {
			msg.Die("Failed to convert a dependency: %v", err)
		}

		msg.Puts("%s\n\n", resource)
	}
}

// BrewResource represents a Homebrew resource definition for a Go dependency.
// See: http://www.rubydoc.info/github/Homebrew/homebrew/master/Resource/Go
type BrewResource struct {
	Name             string
	URL              string
	Revision         string
	DownloadStrategy string
}

// BrewResourceFromLock converts a Glide Lock to a BrewResource
func BrewResourceFromLock(lock *cfg.Lock) (*BrewResource, error) {
	// Get repo info about the locked dependency and convert to homebrew's
	// resource attributes
	dep := cfg.DependencyFromLock(lock)

	repo, err := dep.GetRepo("")
	if err != nil {
		return nil, err
	}

	br := BrewResource{
		Name:             dep.Name,
		URL:              repo.Remote(),
		Revision:         lock.Version,
		DownloadStrategy: string(repo.Vcs()),
	}

	return &br, nil
}

// String serializes a BrewResource into Homebrew's syntax, for inclusion in a
// formula.
func (br *BrewResource) String() string {
	//return fmt.Sprintf("resource \"%s\" do\n%s\nend", br.Name, "")
	return fmt.Sprintf(`go_resource "%s" do
  url "%s", :using => :%s, :revision => "%s"
end`, br.Name, br.URL, br.DownloadStrategy, br.Revision)
}
