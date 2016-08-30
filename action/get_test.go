package action

import (
	"io/ioutil"
	"testing"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
)

func TestAddPkgsToConfig(t *testing.T) {
	// Route output to discard so it's not displayed with the test output.
	o := msg.Default.Stderr
	msg.Default.Stderr = ioutil.Discard

	conf := new(cfg.Config)
	dep := new(cfg.Dependency)
	dep.Name = "github.com/Masterminds/cookoo"
	conf.Imports = append(conf.Imports, dep)

	names := []string{
		"github.com/Masterminds/cookoo/fmt",
		"github.com/Masterminds/semver",
	}

	addPkgsToConfig(conf, names, false, true, false)

	if !conf.HasDependency("github.com/Masterminds/semver") {
		t.Error("addPkgsToConfig failed to add github.com/Masterminds/semver")
	}

	// Restore messaging to original location
	msg.Default.Stderr = o
}
