package cmd

import (
	"fmt"
	"os"

	"github.com/Masterminds/cookoo"
)

func Mkdir(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	target := os.Getenv("GOPATH")
	if len(target) == 0 {
		return nil, fmt.Errorf("$GOPATH appears to be unset.")
	}

	target = fmt.Sprintf("%s/src", target)

	if err := os.MkdirAll(target, os.ModeDir|0755); err != nil {
		return false, fmt.Errorf("Failed to make directory %s: %s", target, err)
	}

	return true, nil
}
