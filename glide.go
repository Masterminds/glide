// Glide is a command line utility that manages Go project dependencies and
// your GOPATH.
//
// Dependencies are managed via a glide.yaml in the root of a project. The yaml
// file lets you specify projects, versions (tags, branches, or references),
// and even alias one location in as other one. Aliasing is useful when supporting
// forks without needing to rewrite the imports in a codebase.
//
// A glide.yaml file looks like:
//
// 		package: github.com/Masterminds/glide
// 		imports:
//			- package: github.com/Masterminds/cookoo
//			  vcs: git
//			  ref: 1.1.0
//			  subpackages: **
//			- package: github.com/kylelemons/go-gypsy
//			  subpackages: yaml
//
// Glide puts dependencies in a _vendor directory. Go utilities require this to
// be in your GOPATH. Glide makes this easy. Use the `glide in` command to enter
// a shell (your default) with the GOPATH set to the projects _vendor directory.
// To leave this shell simply exit it.
//
// If your .bashrc, .zshrc, or other startup shell sets your GOPATH you many need
// to optionally set it using something like:
//
//		if [ "" = "${GOPATH}" ]; then
//		  export GOPATH="/some/dir"
//		fi
//
// For more information use the `glide help` command or see https://github.com/Masterminds/glide
package main

import (
	"github.com/Masterminds/glide/cmd"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/cookoo/cli"

	// Aliasing to ccli as long as cookoo/cli is imported with the same name.
	ccli "github.com/codegangsta/cli"

	"flag"
	"os"
)

var version string = "0.2.0-dev"

const Summary = "Manage Go projects with ease."
const Usage = `Manage dependencies, naming, and GOPATH for your Go projects.

Examples:
	$ glide create
	$ glide in
	$ glide install
	$ glide update
	$ glide rebuild

COMMANDS
========

Utilities:

- status: Print a status report.

Dependency management:

- create: Initialize a new project, creating a template glide.yaml.
- install: Install all packages in the glide.yaml.
- update: Update existing packages (alias: 'up').
- rebuild: Rebuild ('go build') the dependencies.

Project tools:

- into: "glide into /my/project" is the same as running "cd /my/project && glide in"
- gopath: Emits the GOPATH for the current project. Useful for things like
  manually setting GOPATH: GOPATH=$(glide gopath)

Importing:

- godeps: Import Godeps and Godeps-Git files and display the would-be yaml file.

FILES
=====

Each project should have a 'glide.yaml' file in the project directory. Files
look something like this:

	package: github.com/Masterminds/glide
	imports:
		- package: github.com/Masterminds/cookoo
		  vcs: git
		  ref: 1.1.0
		  subpackages: **
  		- package: github.com/kylelemons/go-gypsy
		  subpackages: yaml
`

func main() {
	reg, router, cxt := cookoo.Cookoo()

	routes(reg, cxt)

	app := ccli.NewApp()
	app.Name = "glide"
	app.Usage = Usage
	app.Version = version
	app.Flags = []ccli.Flag{
		ccli.StringFlag{
			Name:  "yaml, y",
			Value: "glide.yaml",
			Usage: "Set a YAML configuration file.",
		},
		ccli.BoolFlag{
			Name:  "quiet, q",
			Usage: "Quiet (no info or debug messages)",
		},
	}

	app.Commands = commands(cxt, router)

	app.Run(os.Args)

	// if err := router.HandleRequest("@startup", cxt, false); err != nil {
	// 	fmt.Printf("Oops! %s\n", err)
	// 	os.Exit(1)
	// }

}

func commands(cxt cookoo.Context, router *cookoo.Router) []ccli.Command {
	return []ccli.Command{
		{
			Name:  "in",
			Usage: "Glide into a commandline shell preconfigured for your project",
			Action: func(c *ccli.Context) {
				cxt.Put("cxt:yaml", c.String("yaml"))
				router.HandleRequest("in", cxt, false)
			},
		},
		{
			Name:      "status",
			ShortName: "s",
			Usage:     "Display a status report",
			Action: func(c *ccli.Context) {
				cxt.Put("cxt:yaml", c.String("yaml"))
				router.HandleRequest("status", cxt, false)
			},
		},
	}
}

func routes(reg *cookoo.Registry, cxt cookoo.Context) {

	flags := flag.NewFlagSet("global", flag.PanicOnError)
	flags.Bool("h", false, "Print help text.")
	flags.Bool("q", false, "Quiet (no info or debug messages)")
	flags.String("yaml", "glide.yaml", "Set a YAML configuration file.")

	cxt.Put("os.Args", os.Args)

	//reg.Route("@startup", "Parse args and send to the right subcommand.").
	// Does(cli.ShiftArgs, "_").Using("n").WithDefault(1).
	// Does(cli.ParseArgs, "remainingArgs").
	// Using("flagset").WithDefault(flags).
	// Using("args").From("cxt:os.Args").
	// Does(cli.ShowHelp, "help").
	// Using("show").From("cxt:h cxt:help").
	// Using("summary").WithDefault(Summary).
	// Using("usage").WithDefault(Usage).
	// Using("flags").WithDefault(flags).
	// Does(cmd.BeQuiet, "quiet").
	// Using("quiet").From("cxt:q").
	// Does(cli.RunSubcommand, "subcommand").
	// Using("default").WithDefault("help").
	// Using("offset").WithDefault(0).
	// Using("args").From("cxt:remainingArgs")

	reg.Route("@ready", "Prepare for glide commands.").
		Does(cmd.ReadyToGlide, "ready").
		Does(cmd.ParseYaml, "cfg").Using("filename").From("cxt:yaml")

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

	reg.Route("out", "Set GOPATH back to former val.").
		Does(cmd.Out, "gopath")

	reg.Route("install", "Install dependencies.").
		Does(cmd.InGopath, "pathIsRight").
		Includes("@ready").
		Does(cmd.Mkdir, "dir").Using("dir").WithDefault("_vendor").
		Does(cmd.LinkPackage, "alias").
		Does(cmd.GetImports, "dependencies").Using("conf").From("cxt:cfg").
		Does(cmd.SetReference, "version").Using("conf").From("cxt:cfg").
		Does(cmd.Rebuild, "rebuild").Using("conf").From("cxt:cfg")

	reg.Route("up", "Update dependencies (alias of 'update')").
		Does(cookoo.ForwardTo, "fwd").Using("route").WithDefault("update")

	reg.Route("update", "Update dependencies.").
		Includes("@ready").
		Does(cmd.CowardMode, "_").
		Does(cmd.UpdateImports, "dependencies").Using("conf").From("cxt:cfg").
		Does(cmd.SetReference, "version").Using("conf").From("cxt:cfg").
		Does(cmd.Rebuild, "rebuild").Using("conf").From("cxt:cfg")

	reg.Route("rebuild", "Rebuild dependencies").
		Includes("@ready").
		Does(cmd.CowardMode, "_").
		Does(cmd.Rebuild, "rebuild").Using("conf").From("cxt:cfg")

	reg.Route("pin", "Print a YAML file with all of the packages pinned to the current version.").
		Includes("@ready").
		Does(cmd.UpdateReferences, "refs").Using("conf").From("cxt:cfg").
		Does(cmd.MergeToYaml, "merged").Using("conf").From("cxt:cfg").
		Does(cmd.WriteYaml, "out").Using("yaml.Node").From("cxt:merged")

	reg.Route("godeps", "Read a Godeps file").
		Includes("@ready").
		Does(cmd.Godeps, "godeps").
		Does(cmd.AddDependencies, "addGodeps").
		Using("dependencies").From("cxt:godeps").
		Using("conf").From("cxt:cfg").
		Does(cmd.GodepsGit, "godepsGit").
		Does(cmd.AddDependencies, "addGodepsGit").
		Using("dependencies").From("cxt:godepsGit").
		Using("conf").From("cxt:cfg").
		// Does(cmd.UpdateReferences, "refs").Using("conf").From("cxt:cfg").
		Does(cmd.MergeToYaml, "merged").Using("conf").From("cxt:cfg").
		Does(cmd.WriteYaml, "out").Using("yaml.Node").From("cxt:merged")

	reg.Route("init", "Initialize Glide (deprecated; use 'create'").
		Does(cookoo.ForwardTo, "fwd").Using("route").WithDefault("create")

	reg.Route("create", "Initialize Glide").
		Does(cmd.InitGlide, "init")

	reg.Route("status", "Status").
		Does(cmd.Status, "status")

	reg.Route("@plugin", "Try to send to a plugin.").
		Includes("@ready").
		Does(cmd.DropToShell, "plugin")
}
