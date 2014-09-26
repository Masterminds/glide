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
	"github.com/codegangsta/cli"

	"os"
)

var version string = "0.2.0-dev"

const Summary = "Manage Go projects with ease."
const Usage = `Manage dependencies, naming, and GOPATH for your Go projects.

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

	app := cli.NewApp()
	app.Name = "glide"
	app.Usage = Usage
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "yaml, y",
			Value: "glide.yaml",
			Usage: "Set a YAML configuration file.",
		},
		cli.BoolFlag{
			Name:  "quiet, q",
			Usage: "Quiet (no info or debug messages)",
		},
	}

	app.Commands = commands(cxt, router)

	app.Run(os.Args)
}

func commands(cxt cookoo.Context, router *cookoo.Router) []cli.Command {
	return []cli.Command{
		{
			Name:      "create",
			ShortName: "init",
			Usage:     "Initialize a new project, creating a template glide.yaml",
			Action: func(c *cli.Context) {
				cxt.Put("q", c.GlobalBool("quiet"))
				cxt.Put("yaml", c.GlobalString("yaml"))
				router.HandleRequest("create", cxt, false)
			},
		},
		{
			Name:  "in",
			Usage: "Glide into a commandline shell preconfigured for your project",
			Action: func(c *cli.Context) {
				cxt.Put("q", c.GlobalBool("quiet"))
				cxt.Put("yaml", c.GlobalString("yaml"))
				router.HandleRequest("in", cxt, false)
			},
		},
		{
			Name:  "install",
			Usage: "Install all packages in the glide.yaml",
			Action: func(c *cli.Context) {
				cxt.Put("q", c.GlobalBool("quiet"))
				cxt.Put("yaml", c.GlobalString("yaml"))
				router.HandleRequest("install", cxt, false)
			},
		},
		{
			Name:  "into",
			Usage: "The same as running \"cd /my/project && glide in\"",
			Action: func(c *cli.Context) {
				cxt.Put("q", c.GlobalBool("quiet"))
				cxt.Put("yaml", c.GlobalString("yaml"))
				cxt.Put("toPath", c.Args()[0])
				router.HandleRequest("into", cxt, false)
			},
		},
		{
			Name:  "godeps",
			Usage: "Import Godeps and Godeps-Git files and display the would-be yaml file",
			Action: func(c *cli.Context) {
				cxt.Put("q", c.GlobalBool("quiet"))
				cxt.Put("yaml", c.GlobalString("yaml"))
				router.HandleRequest("godeps", cxt, false)
			},
		},
		{
			Name:  "gopath",
			Usage: "Display the GOPATH for the present project",
			Description: `Emits the GOPATH for the current project. Useful for
   things like manually setting GOPATH: GOPATH=$(glide gopath)`,
			Action: func(c *cli.Context) {
				cxt.Put("q", c.GlobalBool("quiet"))
				cxt.Put("yaml", c.GlobalString("yaml"))
				router.HandleRequest("gopath", cxt, false)
			},
		},
		{
			Name:  "pin",
			Usage: "Print a YAML file with all of the packages pinned to the current version",
			Action: func(c *cli.Context) {
				cxt.Put("q", c.GlobalBool("quiet"))
				cxt.Put("yaml", c.GlobalString("yaml"))
				router.HandleRequest("pin", cxt, false)
			},
		},
		{
			Name:  "rebuild",
			Usage: "Rebuild ('go build') the dependencies",
			Action: func(c *cli.Context) {
				cxt.Put("q", c.GlobalBool("quiet"))
				cxt.Put("yaml", c.GlobalString("yaml"))
				router.HandleRequest("rebuild", cxt, false)
			},
		},
		{
			Name:      "status",
			ShortName: "s",
			Usage:     "Display a status report",
			Action: func(c *cli.Context) {
				cxt.Put("q", c.GlobalBool("quiet"))
				cxt.Put("yaml", c.GlobalString("yaml"))
				router.HandleRequest("status", cxt, false)
			},
		},
		{
			Name:      "update",
			ShortName: "up",
			Usage:     "Update existing packages",
			Action: func(c *cli.Context) {
				cxt.Put("q", c.GlobalBool("quiet"))
				cxt.Put("yaml", c.GlobalString("yaml"))
				router.HandleRequest("update", cxt, false)
			},
		},
	}
}

func routes(reg *cookoo.Registry, cxt cookoo.Context) {

	reg.Route("@startup", "Parse args and send to the right subcommand.").
		// TODO: Add setup for debug in addition to quiet.
		Does(cmd.BeQuiet, "quiet").
		Using("quiet").From("cxt:q")

	reg.Route("@ready", "Prepare for glide commands.").
		Does(cmd.ReadyToGlide, "ready").
		Does(cmd.ParseYaml, "cfg").Using("filename").From("cxt:yaml")

	reg.Route("into", "Creates a new Glide shell.").
		Includes("@startup").
		Does(cmd.AlreadyGliding, "isGliding").
		Does(cmd.Into, "in").Using("into").From("cxt:toPath").
		Using("into").WithDefault("").From("cxt:toPath").
		Includes("@ready")

	reg.Route("in", "Set GOPATH and supporting env vars.").
		Includes("@startup").
		Does(cmd.AlreadyGliding, "isGliding").
		Includes("@ready").
		Does(cmd.Into, "in").
		Using("into").WithDefault("").From("cxt:toPath").
		Using("conf").From("cxt:cfg")

	reg.Route("gopath", "Return the GOPATH for the present project.").
		Includes("@startup").
		Does(cmd.In, "gopath")

	// reg.Route("out", "Set GOPATH back to former val.").
	// 	Does(cmd.Out, "gopath")

	reg.Route("install", "Install dependencies.").
		Includes("@startup").
		Does(cmd.InGopath, "pathIsRight").
		Includes("@ready").
		Does(cmd.Mkdir, "dir").Using("dir").WithDefault("_vendor").
		Does(cmd.LinkPackage, "alias").
		Does(cmd.GetImports, "dependencies").Using("conf").From("cxt:cfg").
		Does(cmd.SetReference, "version").Using("conf").From("cxt:cfg").
		Does(cmd.Rebuild, "rebuild").Using("conf").From("cxt:cfg")

	reg.Route("update", "Update dependencies.").
		Includes("@startup").
		Includes("@ready").
		Does(cmd.CowardMode, "_").
		Does(cmd.UpdateImports, "dependencies").Using("conf").From("cxt:cfg").
		Does(cmd.SetReference, "version").Using("conf").From("cxt:cfg").
		Does(cmd.Rebuild, "rebuild").Using("conf").From("cxt:cfg")

	reg.Route("rebuild", "Rebuild dependencies").
		Includes("@startup").
		Includes("@ready").
		Does(cmd.CowardMode, "_").
		Does(cmd.Rebuild, "rebuild").Using("conf").From("cxt:cfg")

	reg.Route("pin", "Print a YAML file with all of the packages pinned to the current version.").
		Includes("@startup").
		Includes("@ready").
		Does(cmd.UpdateReferences, "refs").Using("conf").From("cxt:cfg").
		Does(cmd.MergeToYaml, "merged").Using("conf").From("cxt:cfg").
		Does(cmd.WriteYaml, "out").Using("yaml.Node").From("cxt:merged")

	reg.Route("godeps", "Read a Godeps file").
		Includes("@startup").
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

	reg.Route("create", "Initialize Glide").
		Includes("@startup").
		Does(cmd.InitGlide, "init")

	reg.Route("status", "Status").
		Includes("@startup").
		Does(cmd.Status, "status")

	reg.Route("@plugin", "Try to send to a plugin.").
		Includes("@ready").
		Does(cmd.DropToShell, "plugin")
}
