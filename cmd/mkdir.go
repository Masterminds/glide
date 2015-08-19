package cmd

import (
	"fmt"
	"os"

	"github.com/Masterminds/cookoo"
)

// Mkdir creates the src directory within the GOPATH.
func Mkdir(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {

	target := p.Get("dir", "").(string)
	if len(target) == 0 {
		return nil, fmt.Errorf("Vendor path appears to be unset")
	}

	if err := os.MkdirAll(target, os.ModeDir|0755); err != nil {
		return false, fmt.Errorf("Failed to make directory %s: %s", target, err)
	}

	return true, nil
}
