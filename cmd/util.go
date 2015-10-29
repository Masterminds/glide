package cmd

import (
	"fmt"
	"go/build"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/Masterminds/cookoo"
)

// Quiet, when set to true, can suppress Info and Debug messages.
var Quiet = false
var IsDebugging = false
var NoColor = false

// BeQuiet supresses Info and Debug messages.
func BeQuiet(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	Quiet = p.Get("quiet", false).(bool)
	IsDebugging = p.Get("debug", false).(bool)
	return Quiet, nil
}

// CheckColor turns off the colored output (and uses plain text output) for
// logging depending on the value of the "no-color" flag.
func CheckColor(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	NoColor = p.Get("no-color", false).(bool)
	return NoColor, nil
}

// ReadyToGlide fails if the environment is not sufficient for using glide.
//
// Most importantly, it fails if glide.yaml is not present in the current
// working directory.
func ReadyToGlide(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fname := p.Get("filename", "glide.yaml").(string)
	if _, err := os.Stat(fname); err != nil {
		cwd, _ := os.Getwd()
		return false, fmt.Errorf("%s is missing from %s", fname, cwd)
	}
	return true, nil
}

// VersionGuard ensures that the Go version is correct.
func VersionGuard(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	// 6l was removed in 1.5, when vendoring was introduced.
	cmd := exec.Command("go", "tool", "6l")
	var out string
	if _, err := cmd.CombinedOutput(); err == nil {
		Warn("You must install the Go 1.5 or greater toolchain to work with Glide.\n")
	}
	if os.Getenv("GO15VENDOREXPERIMENT") != "1" {
		Warn("To use Glide, you must set GO15VENDOREXPERIMENT=1\n")
	}

	// Verify the setup isn't for the old version of glide. That is, this is
	// no longer assuming the _vendor directory as the GOPATH. Inform of
	// the change.
	if _, err := os.Stat("_vendor/"); err == nil {
		Warn(`Your setup appears to be for the previous version of Glide.
Previously, vendor packages were stored in _vendor/src/ and
_vendor was set as your GOPATH. As of Go 1.5 the go tools
recognize the vendor directory as a location for these
files. Glide has embraced this. Please remove the _vendor
directory or move the _vendor/src/ directory to vendor/.` + "\n")
	}

	return out, nil
}

// CowardMode checks that the environment is setup before continuing on. If not
// setup and error is returned.
func CowardMode(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	gopath := Gopath()
	if gopath == "" {
		return false, fmt.Errorf("No GOPATH is set.\n")
	}

	_, err := os.Stat(path.Join(gopath, "src"))
	if err != nil {
		Error("Could not find %s/src.\n", gopath)
		Info("As of Glide 0.5/Go 1.5, this is required.\n")
		return false, err
	}

	return true, nil
}

// Check if a directory is empty or not.
func isDirectoryEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)

	if err == io.EOF {
		return true, nil
	}

	return false, err
}

// Gopath gets GOPATH from environment and return the most relevant path.
//
// A GOPATH can contain a colon-separated list of paths. This retrieves the
// GOPATH and returns only the FIRST ("most relevant") path.
//
// This should be used carefully. If, for example, you are looking for a package,
// you may be better off using Gopaths.
func Gopath() string {
	gopaths := Gopaths()
	if len(gopaths) == 0 {
		return ""
	}
	return gopaths[0]
}

// Gopaths retrieves the Gopath as a list when there is more than one path
// listed in the Gopath.
func Gopaths() []string {
	p := os.Getenv("GOPATH")
	p = strings.Trim(p, string(filepath.ListSeparator))
	return filepath.SplitList(p)
}

// BuildCtxt is a convenience wrapper for not having to import go/build
// anywhere else
type BuildCtxt struct {
	build.Context
}

// GetBuildContext returns a build context from go/build. When the $GOROOT
// variable is not set in the users environment it sets the context's root
// path to the path returned by 'go env GOROOT'.
func GetBuildContext() (*BuildCtxt, error) {
	buildContext := &BuildCtxt{build.Default}
	if goRoot := os.Getenv("GOROOT"); len(goRoot) == 0 {
		out, err := exec.Command("go", "env", "GOROOT").Output()
		if goRoot = strings.TrimSpace(string(out)); len(goRoot) == 0 || err != nil {
			return nil, fmt.Errorf("Please set the $GOROOT environment " +
				"variable to use this command\n")
		}
		buildContext.GOROOT = goRoot
	}
	return buildContext, nil
}

func fileExist(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
