package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Masterminds/cookoo"
)

// DropToShell executes a glide plugin. A command that's implemented by
// another application is executed in a similar manner to the way git commands
// work. For example, 'glide foo' would try to execute the application glide-foo.
// Params:
//   - command: the name of the command to attempt executing.
func DropToShell(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	args := c.Get("os.Args", nil).([]string)
	command := p.Get("command", "").(string)

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

	cmd := "glide-" + command
	var fullcmd string
	if fullcmd, err = exec.LookPath(cmd); err != nil {
		fullcmd = projpath + "/" + cmd
		if _, err := os.Stat(fullcmd); err != nil {
			return nil, fmt.Errorf("Command %s does not exist.", cmd)
		}
	}

	// Turning os.Args first argument from `glide` to `glide-command`
	args[0] = cmd
	// Removing the first argument (command)
	removed := false
	for i, v := range args {
		if removed == false && v == command {
			args = append(args[:i], args[i+1:]...)
			removed = true
		}
	}
	pa := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir:   cwd,
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
