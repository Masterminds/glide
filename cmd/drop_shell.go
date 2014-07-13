package cmd

import (
	"github.com/Masterminds/cookoo"
	"fmt"
	"os"
	"os/exec"
)

func DropToShell(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	args := c.Get("os.Args", nil).([]string)

	if len(args) == 0 {
		return nil, fmt.Errorf("Could not get os.Args.")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	projpath := cwd
	if tmp := os.Getenv("GLIDE_PROJECT"); len(tmp) != 0 {
		projpath = tmp
	}

	cmd := "glide-" + args[0]
	var fullcmd string
	if fullcmd, err = exec.LookPath(cmd); err != nil {
		fullcmd = projpath + "/" + cmd
		if _, err := os.Stat(fullcmd); err != nil {
			return nil, fmt.Errorf("Command %s does not exist.", cmd)
		}
	}

	args[0] = cmd
	pa := os.ProcAttr {
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir: cwd,
	}

	fmt.Printf("Delegating to plugin %s (%v)\n", fullcmd, args)

	proc, err := os.StartProcess(fullcmd, args, &pa)
	if err != nil {
		return nil, err
	}

	if _, err := proc.Wait(); err != nil {
		return nil, err
	}
	return nil, nil
}
