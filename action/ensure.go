package action

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/Masterminds/glide/cfg"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/util"
)

// EnsureConfig loads and returns a config file.
//
// Any error will cause an immediate exit, with an error printed to Stderr.
func EnsureConfig() *cfg.Config {
	yamlpath, err := gpath.Glide()
	if err != nil {
		msg.ExitCode(2)
		msg.Die("Failed to find %s file in directory tree: %s", gpath.GlideFile, err)
	}

	yml, err := ioutil.ReadFile(yamlpath)
	if err != nil {
		msg.ExitCode(2)
		msg.Die("Failed to load %s: %s", yamlpath, err)
	}
	conf, err := cfg.ConfigFromYaml(yml)
	if err != nil {
		msg.ExitCode(3)
		msg.Die("Failed to parse %s: %s", yamlpath, err)
	}

	b := filepath.Dir(yamlpath)
	buildContext, err := util.GetBuildContext()
	if err != nil {
		msg.Die("Failed to build an import context while ensuring config: %s", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		msg.Err("Unable to get the current working directory")
	} else {
		// Determining a package name requires a relative path
		b, err = filepath.Rel(b, cwd)
		if err == nil {
			name := buildContext.PackageName(b)
			if name != conf.Name {
				msg.Warn("The name listed in the config file (%s) does not match the current location (%s)", conf.Name, name)
			}
		} else {
			msg.Warn("Problem finding the config file path (%s) relative to the current directory (%s): %s", b, cwd, err)
		}
	}

	return conf
}

// EnsureGoVendor ensures that the Go version is correct.
func EnsureGoVendor() bool {
	// 6l was removed in 1.5, when vendoring was introduced.
	cmd := exec.Command(goExecutable(), "tool", "6l")
	if _, err := cmd.CombinedOutput(); err == nil {
		msg.Warn("You must install the Go 1.5 or greater toolchain to work with Glide.\n")
//		os.Exit(1)
		return false
	}

	// Check if this is go15, which requires GO15VENDOREXPERIMENT
	// Any release after go15 does not require that env var.
	cmd = exec.Command(goExecutable(), "version")
	if out, err := cmd.CombinedOutput(); err != nil {
		msg.Err("Error getting version: %s.\n", err)
		os.Exit(1)
	} else if strings.HasPrefix(string(out), "go version 1.5") {
		// This works with 1.5 and 1.6.
		cmd = exec.Command(goExecutable(), "env", "GO15VENDOREXPERIMENT")
		if out, err := cmd.CombinedOutput(); err != nil {
			msg.Warn("Error looking for $GOVENDOREXPERIMENT: %s.\n", err)
//			os.Exit(1)
			return false
		} else if strings.TrimSpace(string(out)) != "1" {
			msg.Warn("To use Glide, you must set GO15VENDOREXPERIMENT=1")
//			os.Exit(1)
			return false
		}
	}

	// In the case where vendoring is explicitly disabled, balk.
	if os.Getenv("GO15VENDOREXPERIMENT") == "0" {
		msg.Warn("To use Glide, you must set GO15VENDOREXPERIMENT=1")
//		os.Exit(1)
		return false
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
//		os.Exit(1)
		return false
	}
	
	return true
}

// EnsureVendorDir ensures that a vendor/ directory is present in the cwd.
func EnsureVendorDir() {
	var vendorDir string
	if gpath.UseGoVendor {
		vendorDir = gpath.VendorDir
	} else {
		vendorDir = gpath.GoPathVendorDir
	}
	fi, err := os.Stat(vendorDir)
	if err != nil {
		msg.Debug("Creating %s", vendorDir)
		if err := os.MkdirAll(vendorDir, os.ModeDir|0755); err != nil {
			msg.Die("Could not create %s: %s", vendorDir, err)
		}
	} else if !fi.IsDir() {
		msg.Die("Vendor is not a directory")
	}
	
	if !gpath.UseGoVendor {
		// create symblic link to gpath.GoPathVendorDir to make glide work as normal
		fi, err := os.Lstat(gpath.VendorDir)
		if err != nil || (fi.Mode() & os.ModeSymlink) != 0 {
			var symlinkOk = false
			if err == nil && (fi.Mode() & os.ModeSymlink) != 0 {
				if p, err := os.Readlink(gpath.VendorDir); err == nil && p == gpath.GoPathVendorDir {
					symlinkOk = true
				} else {
					msg.Debug("Remove symlink %s -> %s", gpath.VendorDir, p)
					if err := os.Remove(gpath.VendorDir); err != nil {
						msg.Err("Could not remove symlink %s: %s", gpath.VendorDir, err)
					}
				}
			}
			if !symlinkOk {
				msg.Debug("Creating symlink %s -> %s", gpath.VendorDir, gpath.GoPathVendorDir)
				if err := os.Symlink(gpath.GoPathVendorDir, gpath.VendorDir); err != nil {
					msg.Die("Could not create symlink %s: %s", gpath.VendorDir, err)
				}
			}
		} else {
			msg.Die("%s exists but not a symlink", gpath.VendorDir)
		}
	}
}

// EnsureGopath fails if GOPATH is not set, or if $GOPATH/src is missing.
//
// Otherwise it returns the value of GOPATH.
func EnsureGopath() string {
	gps := gpath.Gopaths()
	if len(gps) == 0 {
		msg.Die("$GOPATH is not set.")
	}

	for _, gp := range gps {
		_, err := os.Stat(path.Join(gp, "src"))
		if err != nil {
			msg.Warn("%s", err)
			continue
		}
		return gp
	}

	msg.Err("Could not find any of %s/src.\n", strings.Join(gps, "/src, "))
	msg.Info("As of Glide 0.5/Go 1.5, this is required.\n")
	msg.Die("Without src, cannot continue.")
	return ""
}

// goExecutable checks for a set environment variable of GLIDE_GO_EXECUTABLE
// for the go executable name. The Google App Engine SDK ships with a python
// wrapper called goapp
//
// Example usage: GLIDE_GO_EXECUTABLE=goapp glide install
func goExecutable() string {
	goExecutable := os.Getenv("GLIDE_GO_EXECUTABLE")
	if len(goExecutable) <= 0 {
		goExecutable = "go"
	}

	return goExecutable
}
