package cmd

import (
	"os"

	"github.com/Masterminds/cookoo"
)

// Status is a command that prints the status of the glide and expected gopath.
func Status(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {

	vpath, err := VendorPath(c)
	if err != nil {
		Error("Could not get vendor path: %s", err)
	}

	gopath := os.Getenv("GOPATH")

	Info("Vendor path is: %s\n", vpath)
	Info("GOPATH is: %s\n", gopath)

	stat, err := os.Stat(vpath)
	if err != nil {
		Error("Error with vendor path: %s\n", err)
		Info("Did you forget to do a `glide init`?\n")
		return false, nil
	}
	if !stat.IsDir() {
		Error("vendir is not a directory.\n")
		return false, nil
	}

	return true, nil
}
