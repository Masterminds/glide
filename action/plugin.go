package action

import (
	"os"
	"os/exec"

	"github.com/Masterminds/glide/msg"
)

// Plugin attempts to find and execute a plugin based on a command.
//
// Exit code 99 means the plugin was never executed. Code 1 means the program
// exited badly.
func Plugin(command string, args []string) {

	cwd, err := os.Getwd()
	if err != nil {
		msg.ExitCode(99)
		msg.Die("Could not get working directory: %s", err)
	}

	cmd := "glide-" + command
	var fullcmd string
	if fullcmd, err = exec.LookPath(cmd); err != nil {
		fullcmd = cwd + "/" + cmd
		if _, err := os.Stat(fullcmd); err != nil {
			msg.ExitCode(99)
			msg.Die("Command %s does not exist.", cmd)
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

	msg.Debug("Delegating to plugin %s (%v)\n", fullcmd, args)

	proc, err := os.StartProcess(fullcmd, args, &pa)
	if err != nil {
		msg.Err("Failed to execute %s: %s", cmd, err)
		os.Exit(98)
	}

	if _, err := proc.Wait(); err != nil {
		msg.Err(err.Error())
		os.Exit(1)
	}
}
