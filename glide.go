package main

import (
	"github.com/technosophos/glide/cmd"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/cookoo/cli"

	"flag"
	"fmt"
	"os"
)

const Summary = "Manage Go projects with ease."
const Usage = `Manage dependencies, naming, and GOPATH for your Go projects.

Examples:
	$ glide init
	$ glide in
	$ glide install

COMMANDS
========

- help: Show this help message (alias of -h)

- init: Initialize a new project, creating a template glide.yaml
- install: Install all packages in the glide.yaml
- update: Update existing packages

- in: Glide into a commandline shell preconfigured for your project (with
  GOPATH set).
- into: "glide into /my/project" is the same as running "cd /my/project && glide in"
- gopath: Emits the GOPATH for the current project. Useful for things like
  manually setting GOPATH: GOPATH=$(glide gopath)

`

func main() {
	reg, router, cxt := cookoo.Cookoo()

	routes(reg, cxt)

	if err := router.HandleRequest("@startup", cxt, false); err != nil {
		fmt.Printf("Starup error: %s\n", err)
		os.Exit(1)
	}

	next := cxt.Get("subcommand", "help").(string)
	if router.HasRoute(next) {
		if err := router.HandleRequest(next, cxt, false); err != nil {
			fmt.Printf("Oops! %s\n", err)
			os.Exit(1)
		}
	} else {
		if err := router.HandleRequest("@plugin", cxt, false); err != nil {
			fmt.Printf("Oops! %s\n", err)
			os.Exit(1)
		}
	}

}

func routes(reg *cookoo.Registry, cxt cookoo.Context) {

	flags := flag.NewFlagSet("global", flag.PanicOnError)
	flags.Bool("h", false, "Print help text.")

	cxt.Put("os.Args", os.Args)

	reg.Route("@startup", "Parse args and send to the right subcommand.").
		Does(cli.ShiftArgs, "_").Using("n").WithDefault(1).
		Does(cli.ParseArgs, "parseargs").
		Using("flagset").WithDefault(flags).
		Using("args").From("cxt:os.Args").
		Does(cli.ShowHelp, "help").
		Using("show").From("cxt:h cxt:help").
		Using("summary").WithDefault(Summary).
		Using("usage").WithDefault(Usage).
		Using("flags").WithDefault(flags).
		Does(subcommand, "subcommand").
		Using("args").From("cxt:os.Args")

	reg.Route("@ready", "Prepare for glide commands.").
		Does(cmd.ReadyToGlide, "ready").
		Does(cmd.ParseYaml, "cfg")

	reg.Route("help", "Print help.").
		Does(cli.ShowHelp, "help").
		Using("show").WithDefault(true).
		Using("summary").WithDefault(Summary).
		Using("usage").WithDefault(Usage).
		Using("flags").WithDefault(flags)

	reg.Route("into", "Creates a new Glide shell.").
		Does(cmd.AlreadyGliding, "isGliding").
		Does(cli.ShiftArgs, "toPath").Using("n").WithDefault(2).
		Does(cmd.Into, "in").Using("into").From("cxt:toPath").
		Using("into").WithDefault("").From("cxt:toPath").
		Includes("@ready")

	reg.Route("in", "Set GOPATH and supporting env vars.").
		Does(cmd.AlreadyGliding, "isGliding").
		Includes("@ready").
		//Does(cli.ShiftArgs, "toPath").Using("n").WithDefault(1).
		Does(cmd.Into, "in").
		Using("into").WithDefault("").From("cxt:toPath").
		Using("conf").From("cxt:cfg")

	reg.Route("gopath", "Return the GOPATH for the present project.").
		Does(cmd.In, "gopath")
	/*
		Includes("@ready").
		Does(cmd.AlreadyGliding, "isGliding").
		Does(cmd.In, "gopath")
	*/
	reg.Route("out", "Set GOPATH back to former val.").
		Does(cmd.Out, "gopath")
	reg.Route("install", "Install dependencies.").
		Includes("@ready").
		Does(cmd.Mkdir, "dir").Using("dir").WithDefault("_vendor").
		Does(cmd.LinkPackage, "alias").
		Does(cmd.GetImports, "dependencies").Using("conf").From("cxt:cfg").
		Does(cmd.SetReference, "version").Using("conf").From("cxt:cfg")

	reg.Route("update", "Update dependencies.").
		Includes("@ready").
		Does(cmd.CowardMode, "_").
		Does(cmd.UpdateImports, "dependencies").Using("conf").From("cxt:cfg").
		Does(cmd.SetReference, "version").Using("conf").From("cxt:cfg")

	reg.Route("init", "Initialize Glide").
		Does(cmd.InitGlide, "init")

	reg.Route("@plugin", "Try to send to a plugin.").
		Includes("@ready").
		Does(func (c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
			fmt.Printf("Command '%s' not found.", c.Get("subcommand", "").(string))
			return nil, nil
		}, "_")
}

func subcommand(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	args := p.Get("args", []string{"help"}).([]string)
	if len(args) == 0 {
		return "help", nil
	}
	return args[0], nil
}
