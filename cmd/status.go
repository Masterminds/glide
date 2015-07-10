package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/Masterminds/cookoo"
)

// Status is a command that prints the status of the glide and expected gopath.
func Status(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	if os.Getenv("ALREADY_GLIDING") == "1" {
		fmt.Println("glide in: true")
	} else {
		fmt.Println("glide in: false")
	}

	cwd, _ := os.Getwd()
	gopath := os.Getenv("GOPATH")
	vendor := c.Get("VendorDir", "vendor").(string)

	expected := path.Join(cwd, vendor)
	if gopath != expected {
		fmt.Println("gopath: unexpected")
	} else {
		fmt.Println("gopath: ok")
	}
	return true, nil
}
