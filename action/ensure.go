package action

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
)

// EnsureConfig loads and returns a config file.
//
// Any error will cause an immediate exit, with an error printed to Stderr.
func EnsureConfig(yamlpath string) *cfg.Config {
	yml, err := ioutil.ReadFile(yamlpath)
	if err != nil {
		msg.Error("Failed to load %s: %s", yamlpath, err)
		os.Exit(2)
	}
	conf, err := cfg.ConfigFromYaml(yml)
	if err != nil {
		msg.Error("Failed to parse %s: %s", yamlpath, err)
		os.Exit(3)
	}

	return conf
}

func EnsureCacheDir() {
	msg.Warn("ensure.go: ensureCacheDir is not implemented.")
}

// EnsureGoVendor ensures that the Go version is correct.
func EnsureGoVendor() {
	// 6l was removed in 1.5, when vendoring was introduced.
	cmd := exec.Command("go", "tool", "6l")
	if _, err := cmd.CombinedOutput(); err == nil {
		msg.Warn("You must install the Go 1.5 or greater toolchain to work with Glide.\n")
		os.Exit(1)
	}
	if os.Getenv("GO15VENDOREXPERIMENT") != "1" {
		msg.Warn("To use Glide, you must set GO15VENDOREXPERIMENT=1\n")
		os.Exit(1)
	}

	// Verify the setup isn't for the old version of glide. That is, this is
	// no longer assuming the _vendor directory as the GOPATH. Inform of
	// the change.
	if _, err := os.Stat("_vendor/"); err == nil {
		msg.Warn(`Your setup appears to be for the previous version of Glide.
Previously, vendor packages were stored in _vendor/src/ and
_vendor was set as your GOPATH. As of Go 1.5 the go tools
recognize the vendor directory as a location for these
files. Glide has embraced this. Please remove the _vendor
directory or move the _vendor/src/ directory to vendor/.` + "\n")
		os.Exit(1)
	}
}
