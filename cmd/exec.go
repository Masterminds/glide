package cmd

import (
	"fmt"
	"github.com/Masterminds/cookoo"
	"github.com/codegangsta/cli"
	"os"
	"os/exec"
)

func ExecCmd(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	args := p.Get("args", nil).(cli.Args)
	fname := p.Get("filename", "glide.yaml").(string)

	if len(args) == 0 {
		return nil, fmt.Errorf("No command to execute")
	}

	gopath, err := GlideGopath(fname)
	if err != nil {
		return false, err
	}

	err = os.Setenv("GOPATH", gopath)
	if err != nil {
		return false, err
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	return true, nil
}
