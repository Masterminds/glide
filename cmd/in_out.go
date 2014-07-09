package cmd

import (
	"github.com/Masterminds/cookoo"
	"fmt"
	"os"
)

func In(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	gopath := fmt.Sprintf("%s/_vendor", cwd)
	binpath := fmt.Sprintf("$PATH:%s/bin", gopath)
	old_gopath := os.Getenv("GOPATH")
	old_binpath := os.Getenv("PATH")

	os.Setenv("OLD_GOPATH", old_gopath)
	os.Setenv("OLD_PATH", old_binpath)
	os.Setenv("GOPATH", gopath)
	os.Setenv("PATH_TEST", binpath)
	fmt.Printf("export GOPATH=%s\n", gopath)

	return nil, nil
}

func Out(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	os.Setenv("GOPATH", os.Getenv("OLD_GOPATH"))
	os.Setenv("PATH_TEST", os.Getenv("OLD_PATH"))
	return true, nil
}
