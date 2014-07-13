package cmd

import (
	"github.com/Masterminds/cookoo"
	"fmt"
	"os"
)

func Status(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	if os.Getenv("ALREADY_GLIDING") == "1" {
		fmt.Println("glide in: true")
	} else {
		fmt.Println("glide in: false")
	}

	cwd, _ := os.Getwd()
	gopath := os.Getenv("GOPATH")

	expected := fmt.Sprintf("%s/_vendor", cwd)
	if gopath != expected {
		fmt.Println("gopath: unexpected")
	} else {
		fmt.Println("gopath: ok")
	}
	return true, nil
}
