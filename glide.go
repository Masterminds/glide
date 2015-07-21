// Glide is a command line utility that manages Go project dependencies and
// your GOPATH.
//
// Dependencies are managed via a glide.yaml in the root of a project. The yaml
//
// Params:
// 	- filename (string): The name of the glide YAML file. Default is glide.yaml.
// 	- project (string): The name of the project. Default is 'main'.
// file lets you specify projects, versions (tags, branches, or references),
// and even alias one location in as other one. Aliasing is useful when supporting
// forks without needing to rewrite the imports in a codebase.
//
// A glide.yaml file looks like:
//
//		package: github.com/Masterminds/glide
//		imports:
//			- package: github.com/Masterminds/cookoo
//			  vcs: git
//			  ref: 1.1.0
//			  subpackages: **
//			- package: github.com/kylelemons/go-gypsy
//			  subpackages: yaml
//
// Glide puts dependencies in a vendor directory. Go utilities require this to
// be in your GOPATH. Glide makes this easy. Use the `glide in` command to enter
// a shell (your default) with the GOPATH set to the projects vendor directory.
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

	"fmt"
	"os"
)

var version = "0.5.0-dev"

const usage = `Manage dependencies, naming, and GOPATH for your Go projects.

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

NOTE: As of Glide 0.5, the commands 'in', 'into', 'gopath', and 'instal' no
longer exist.
`

var VendorDir = "vendor"

func main() {
	reg, router, cxt := cookoo.Cookoo()
	cxt.Put("VendorDir", VendorDir)

	routes(reg, cxt)

	app := cli.NewApp()
	app.Name = "glide"
	app.Usage = usage
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
	app.CommandNotFound = func(c *cli.Context, command string) {
		cxt.Put("os.Args", os.Args)
		cxt.Put("command", command)
		setupHandler(c, "@plugin", cxt, router)
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
			Description: `This command starts from a project without Glide and
	sets it up. Once this step is done, you may edit the glide.yaml file and then
	you may run 'glide install' to fetch your initial dependencies.

	By default, the project name is 'main'. You can specify an alternative on
	the commandline:

		$ glide create github.com/Masterminds/foo

	For a project that already has a glide.yaml file, you may skip 'glide create'
	and instead run 'glide up'.`,
			Action: func(c *cli.Context) {
				if len(c.Args()) >= 1 {
					cxt.Put("project", c.Args()[0])
				}
				setupHandler(c, "create", cxt, router)
			},
		},
		{
			Name:  "get",
			Usage: "Run 'go get' and update the glide.yaml file with the new package.",
			Description: `Gets the package using 'go get' and then adds that file
	to the glide.yaml file.

		$ glide get github.com/Masterminds/cookoo/web

	The above will install the package github.com/Masterminds/cookoo and add
	the subpackage 'web'.`,
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					fmt.Println("Oops! Package name is required.")
					os.Exit(1)
				}
				cxt.Put("package", c.Args()[0])
				setupHandler(c, "get", cxt, router)
			},
		},
		{
			Name:  "import",
			Usage: "Import files from other dependency management systems.",
			Subcommands: []cli.Command{
				{
					Name:  "godeps",
					Usage: "Import Godep's Godeps.json files and display the would-be yaml file",
					Action: func(c *cli.Context) {
						setupHandler(c, "import godep", cxt, router)
					},
				},
				{
					Name:  "gpm",
					Usage: "Import GPM's Godeps and Godeps-Git files and display the would-be yaml file",
					Action: func(c *cli.Context) {
						setupHandler(c, "import gpm", cxt, router)
					},
				},
			},
		},
		{
			Name:      "env",
			ShortName: "gopath",
			Usage:     "Display environment variables for the present project",
			Description: `Emits the environment for the current project. Useful for
   things like manually setting GOPATH: GOPATH=$(glide gopath)`,
			Action: func(c *cli.Context) {
				setupHandler(c, "env", cxt, router)
			},
		},
		{
			Name:  "pin",
			Usage: "Print a YAML file with all of the packages pinned to the current version",
			Description: `Begins with the current glide.yaml and sets an absolute ref
    for every package. The version is derived from the repository version. It will be
    either a commit or a tag, depending on the state of the VCS tree.

    By default, output is written to standard out. However, if you supply a filename,
    the data will be written to that:

        $ glide pin glide.yaml

    The above will overwrite your glide.yaml file. You have been warned.
	`,
			Action: func(c *cli.Context) {
				outfile := ""
				if len(c.Args()) == 1 {
					outfile = c.Args()[0]
				}
				cxt.Put("toPath", outfile)
				setupHandler(c, "pin", cxt, router)
			},
		},
		{
			Name:  "rebuild",
			Usage: "Rebuild ('go build') the dependencies",
			Description: `This rebuilds the packages' '.a' files. On some systems
	this can improve performance on subsequent 'go run' and 'go build' calls.`,
			Action: func(c *cli.Context) {
				setupHandler(c, "rebuild", cxt, router)
			},
		},
		{
			Name:      "status",
			ShortName: "s",
			Usage:     "Display a status report",
			Action: func(c *cli.Context) {
				setupHandler(c, "status", cxt, router)
			},
		},
		{
			Name:      "update",
			ShortName: "up",
			Usage:     "Update a project's dependencies",
			Description: `This uses the native VCS of each package to try to
	pull the most applicable updates. Packages with fixed refs (Versions or
	tags) will not be updated. Packages with no ref or with a branch ref will
	be updated as expected.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "preserve-packages",
					Usage: "Set to keep unspecified vendor packages.",
				},
			},
			Action: func(c *cli.Context) {
				cxt.Put("deleteOptOut", c.Bool("preserve-packages"))
				setupHandler(c, "update", cxt, router)
			},
		},
		{
			Name:  "guess",
			Usage: "Guess dependencies for existing source.",
			Description: `This looks through existing source and dependencies,
	and tries to guess all of the dependent packages.

	By default, 'glide guess' writes to standard output. But if a filename
	is supplied, the results are written to the file:

		$ glide guess glide.yaml

	The above will overwrite the glide.yaml file.`,
			Action: func(c *cli.Context) {
				outfile := ""
				if len(c.Args()) == 1 {
					outfile = c.Args()[0]
				}
				cxt.Put("toPath", outfile)
				setupHandler(c, "guess", cxt, router)
			},
		},
		{
			Name:  "about",
			Usage: "Learn about Glide",
			Action: func(c *cli.Context) {
				setupHandler(c, "about", cxt, router)
			},
		},
	}
}

func setupHandler(c *cli.Context, route string, cxt cookoo.Context, router *cookoo.Router) {
	cxt.Put("q", c.GlobalBool("quiet"))
	cxt.Put("yaml", c.GlobalString("yaml"))
	cxt.Put("cliArgs", c.Args())
	if err := router.HandleRequest(route, cxt, false); err != nil {
		fmt.Printf("Oops! %s\n", err)
		os.Exit(1)
	}
}

func routes(reg *cookoo.Registry, cxt cookoo.Context) {
	reg.Route("@startup", "Parse args and send to the right subcommand.").
		// TODO: Add setup for debug in addition to quiet.
		Does(cmd.BeQuiet, "quiet").
		Using("quiet").From("cxt:q").
		Does(cmd.VersionGuard, "v")

	reg.Route("@ready", "Prepare for glide commands.").
		Does(cmd.ReadyToGlide, "ready").Using("filename").From("cxt:yaml").
		Does(cmd.ParseYaml, "cfg").Using("filename").From("cxt:yaml")

	reg.Route("get", "Install a pkg in vendor, and store the results in the glide.yaml").
		Includes("@startup").
		Includes("@ready").
		Does(cmd.Get, "goget").
		Using("filename").From("cxt:yaml").
		Using("package").From("cxt:package").
		Using("conf").From("cxt:cfg").
		Does(cmd.MergeToYaml, "merged").Using("conf").From("cxt:cfg").
		Does(cmd.WriteYaml, "out").
		Using("yaml.Node").From("cxt:merged").
		Using("filename").WithDefault("glide.yaml").From("cxt:yaml")

	reg.Route("exec", "Execute command with GOPATH set.").
		Includes("@startup").
		Includes("@ready").
		Does(cmd.ExecCmd, "cmd").
		Using("args").From("cxt:cliArgs").
		Using("filename").From("cxt:yaml")

	reg.Route("update", "Update dependencies.").
		Includes("@startup").
		Includes("@ready").
		Does(cmd.CowardMode, "_").
		Does(cmd.Mkdir, "dir").Using("dir").WithDefault(VendorDir).
		//Does(cmd.DeleteUnusedPackages, "deleted").
		//Using("conf").From("cxt:cfg").
		//Using("optOut").From("cxt:deleteOptOut").
		Does(cmd.UpdateImports, "dependencies").Using("conf").From("cxt:cfg").
		Does(cmd.SetReference, "version").Using("conf").From("cxt:cfg")
	//Does(cmd.Rebuild, "rebuild").Using("conf").From("cxt:cfg")

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
		Does(cmd.WriteYaml, "out").
		Using("yaml.Node").From("cxt:merged").
		Using("filename").From("cxt:toPath")

	reg.Route("import gpm", "Read a Godeps file").
		Includes("@startup").
		Includes("@ready").
		Does(cmd.GPMGodeps, "godeps").
		Does(cmd.AddDependencies, "addGodeps").
		Using("dependencies").From("cxt:godeps").
		Using("conf").From("cxt:cfg").
		Does(cmd.GPMGodepsGit, "godepsGit").
		Does(cmd.AddDependencies, "addGodepsGit").
		Using("dependencies").From("cxt:godepsGit").
		Using("conf").From("cxt:cfg").
		// Does(cmd.UpdateReferences, "refs").Using("conf").From("cxt:cfg").
		Does(cmd.MergeToYaml, "merged").Using("conf").From("cxt:cfg").
		Does(cmd.WriteYaml, "out").Using("yaml.Node").From("cxt:merged")

	reg.Route("import godep", "Read a Godeps.json file").
		Includes("@startup").
		Includes("@ready").
		Does(cmd.ParseGodepGodeps, "godeps").
		Does(cmd.AddDependencies, "addGodeps").
		Using("dependencies").From("cxt:godeps").
		Using("conf").From("cxt:cfg").
		// Does(cmd.UpdateReferences, "refs").Using("conf").From("cxt:cfg").
		Does(cmd.MergeToYaml, "merged").Using("conf").From("cxt:cfg").
		Does(cmd.WriteYaml, "out").Using("yaml.Node").From("cxt:merged")

	reg.Route("guess", "Guess dependencies").
		Includes("@ready").
		Does(cmd.GuessDeps, "cfg").
		Does(cmd.MergeToYaml, "merged").Using("conf").From("cxt:cfg").
		Does(cmd.WriteYaml, "out").
		Using("yaml.Node").From("cxt:merged").
		Using("filename").From("cxt:toPath")

	reg.Route("create", "Initialize Glide").
		Includes("@startup").
		Does(cmd.InitGlide, "init").
		Using("filename").From("cxt:yaml").
		Using("project").From("cxt:project").WithDefault("main")

	reg.Route("env", "Print environment").
		Includes("@startup").
		Does(cmd.Status, "status")

	reg.Route("status", "Status").
		Includes("@startup").
		Does(cmd.Status, "status")

	reg.Route("about", "Status").
		Includes("@startup").
		Does(cmd.About, "about")

	reg.Route("@plugin", "Try to send to a plugin.").
		Includes("@ready").
		Does(cmd.DropToShell, "plugin").
		Using("command").From("cxt:command")
}
