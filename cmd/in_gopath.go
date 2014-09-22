package cmd

import (
	"github.com/Masterminds/cookoo"

	"path/filepath"
	"fmt"
	"os"
)

func InGopath(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	// Get current dir
	cwd, err := os.Getwd()
	if err != nil {
		return false, err
	}
	cwd = filepath.Join(cwd, "_vendor")
	// Get GOPATH
	gopath, err := filepath.Abs(os.Getenv("GOPATH"))
	if err != nil {
		return false, err
	}

	// Check that they are equal.
	if cwd != gopath {
		Error("For Glide to create a managed _vendor directory, you must set your GOPATH to %s.\n", cwd)
		Info("You can use `glide in` to set GOPATH for you.\n")
		Info("If you are using an external GOPATH, skip to `glide update`.\n")
		return false, fmt.Errorf("GOPATH is %s, but current directory is %s", gopath, cwd)
	}


	return true, nil
}
