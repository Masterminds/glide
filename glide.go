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
//		package: github.com/Masterminds/glide
//		imports:
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

	"fmt"
	"os"
)

var version = "0.4.0-dev"

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
`

func main() {
	reg, router, cxt := cookoo.Cookoo()

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
			Action: func(c *cli.Context) {
				setupHandler(c, "create", cxt, router)
			},
		},
		{
			Name:  "in",
			Usage: "Glide into a commandline shell preconfigured for your project",
			Description: `This is roughly the same as starting a new shell and
	then running GOPATH=$(glide gopath).`,
			Action: func(c *cli.Context) {
				setupHandler(c, "in", cxt, router)
			},
		},
		{
			Name:  "install",
			Usage: "Install all packages in the glide.yaml",
			Action: func(c *cli.Context) {
				setupHandler(c, "install", cxt, router)
			},
		},
		{
			Name:  "into",
			Usage: "The same as running \"cd /my/project && glide in\"",
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					fmt.Println("Oops! directory name is required.")
					os.Exit(1)
				}
				cxt.Put("toPath", c.Args()[0])
				setupHandler(c, "into", cxt, router)
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
			Name:  "godeps",
			Usage: "DEPRECATED. Use `import gpm`. Import Godeps and Godeps-Git files and display the would-be yaml file",
			Action: func(c *cli.Context) {
				setupHandler(c, "import gpm", cxt, router)
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
			Name:  "gopath",
			Usage: "Display the GOPATH for the present project",
			Description: `Emits the GOPATH for the current project. Useful for
   things like manually setting GOPATH: GOPATH=$(glide gopath)`,
			Action: func(c *cli.Context) {
				setupHandler(c, "gopath", cxt, router)
			},
		},
		{
			Name:  "exec",
			Usage: "Execute a command with the Go environment setup",
			Description: `Execute a command inside of the GOPATH. Some commands
    (notably 'go cover') expect themselves to be run from a particular place
    within the GOPATH. This command sets up the environment for such tools.

        $ glide exec go cover

    Most Go tools do not need this.`,
			SkipFlagParsing: true,
			Action: func(c *cli.Context) {
				setupHandler(c, "exec", cxt, router)
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
			Usage:     "Update existing packages",
			Description: `This uses the native VCS of each package to try to
	pull the most applicable updates. Packages with fixed refs (Versions or
	tags) will not be updated. Packages with no ref or with a branch ref will
	be updated as expected.`,
			Action: func(c *cli.Context) {
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
		Using("quiet").From("cxt:q")

	reg.Route("@ready", "Prepare for glide commands.").
		Does(cmd.ReadyToGlide, "ready").Using("filename").From("cxt:yaml").
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
		Does(cmd.In, "gopath").Using("filename").From("cxt:yaml")

	reg.Route("get", "Run 'go get' and install the results in the glide.yaml").
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
		Does(cmd.InitGlide, "init").Using("filename").From("cxt:yaml")

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
