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
	"path/filepath"

	"github.com/Masterminds/glide/cmd"

	"github.com/Masterminds/cookoo"
	"github.com/codegangsta/cli"

	"fmt"
	"os"
	"os/user"
)

var version = "dev"

const usage = `The lightweight vendor package manager for your Go projects.

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
			flatten: true

NOTE: As of Glide 0.5, the commands 'in', 'into', 'gopath', 'status', and 'env'
no longer exist.
`

// VendorDir default vendor directory name
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
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Print Debug messages (verbose)",
		},
		cli.StringFlag{
			Name:   "home",
			Value:  defaultGlideDir(),
			Usage:  "The location of Glide files",
			EnvVar: "GLIDE_HOME",
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
			Usage: "Install one or more package into `vendor/` and add dependency to glide.yaml.",
			Description: `Gets one or more package (like 'go get') and then adds that file
	to the glide.yaml file. Multiple package names can be specified on one line.

		$ glide get github.com/Masterminds/cookoo/web

	The above will install the project github.com/Masterminds/cookoo and add
	the subpackage 'web'.

	If a fetched dependency has a glide.yaml file, 'get' will also install
	all of the dependencies for that dependency. Those are installed in a scoped
	vendor directory. So dependency vendor/foo/bar has its dependencies stored
	in vendor/foo/bar/vendor. This behavior can be disabled using
	'--no-recursive'

	If '--import' is set, this will also read the dependency projects, looking
	for gb, Godep and GPM files. When it finds them, it will build a comparable
	glide.yaml file, and then fetch all of the necessary dependencies. The
	dependencies are then vendored in the appropriate project. Subsequent calls
	to 'glide up' will use the glide.yaml to maintain those dependencies.
	However, only if you call 'glide up --import' will the glide file be
	rebuilt. When '--no-recursive' is used, '--import' does nothing.
	`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "no-recursive",
					Usage: "Disable updating dependencies' dependencies.",
				},
				cli.BoolFlag{
					Name:  "import",
					Usage: "When fetching dependencies, convert Godeps (GPM, Godep) to glide.yaml and pull dependencies",
				},
				cli.BoolFlag{
					Name:  "insecure",
					Usage: "Use http:// rather than https:// to retrieve pacakges.",
				},
			},
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					fmt.Println("Oops! Package name is required.")
					os.Exit(1)
				}
				cxt.Put("packages", []string(c.Args()))
				cxt.Put("skipFlatten", !c.Bool("no-recursive"))
				cxt.Put("insecure", c.Bool("insecure"))
				// FIXME: Are these used anywhere?
				if c.Bool("import") {
					cxt.Put("importGodeps", true)
					cxt.Put("importGPM", true)
					cxt.Put("importGb", true)
				}
				setupHandler(c, "get", cxt, router)
			},
		},
		{
			Name:  "import",
			Usage: "Import files from other dependency management systems.",
			Subcommands: []cli.Command{
				{
					Name:  "godep",
					Usage: "Import Godep's Godeps.json files and display the would-be yaml file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "file, f",
							Usage: "Save all of the discovered dependencies to a Glide YAML file.",
						},
					},
					Action: func(c *cli.Context) {
						cxt.Put("toPath", c.String("file"))
						setupHandler(c, "import godep", cxt, router)
					},
				},
				{
					Name:  "gpm",
					Usage: "Import GPM's Godeps and Godeps-Git files and display the would-be yaml file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "file, f",
							Usage: "Save all of the discovered dependencies to a Glide YAML file.",
						},
					},
					Action: func(c *cli.Context) {
						cxt.Put("toPath", c.String("file"))
						setupHandler(c, "import gpm", cxt, router)
					},
				},
				{
					Name:  "gb",
					Usage: "Import gb's manifest file and display the would-be yaml file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "file, f",
							Usage: "Save all of the discovered dependencies to a Glide YAML file.",
						},
					},
					Action: func(c *cli.Context) {
						cxt.Put("toPath", c.String("file"))
						setupHandler(c, "import gb", cxt, router)
					},
				},
			},
		},
		{
			Name:        "name",
			Usage:       "Print the name of this project.",
			Description: `Read the glide.yaml file and print the name given on the 'package' line.`,
			Action: func(c *cli.Context) {
				setupHandler(c, "name", cxt, router)
			},
		},
		{
			Name:      "novendor",
			ShortName: "nv",
			Usage:     "List all non-vendor paths in a directory.",
			Description: `Given a directory, list all the relevant Go paths that are not vendored.

Example:

			$ go test $(glide novendor)
`,
			Action: func(c *cli.Context) {
				setupHandler(c, "nv", cxt, router)
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
			Name:      "update",
			ShortName: "up",
			Aliases:   []string{"install"},
			Usage:     "Update or install a project's dependencies",
			Description: `This uses the native VCS of each package to try to
	pull the most applicable updates. Packages with fixed refs (Versions or
	tags) will not be updated. Packages with no ref or with a branch ref will
	be updated as expected.

	If a dependency has a glide.yaml file, update will read that file and
	update those dependencies accordingly. Those dependencies are maintained in
	a scoped vendor directory. 'vendor/foo/bar' will have its dependencies
	stored in 'vendor/foo/bar/vendor'. This behavior can be disabled with
	'--no-recursive'.

	Glide will also import Godep, GB, and GPM files as it finds them in dependencies.
	It will create a glide.yaml file from the Godeps data, and then update. This
	has no effect if '--no-recursive' is set.

	If the '--update-vendored' flag (aliased to '-u') is present vendored
	dependencies, stored in your projects VCS repository, will be updated. This
	works by removing the old package, checking out an the repo and setting the
	version, and removing the VCS directory.

	By default, packages that are discovered are considered transient, and are
	not stored in the glide.yaml file. The --file=NAME.yaml flag allows you
	to save the discovered dependencies to a YAML file.
	`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "delete",
					Usage: "Delete vendor packages not specified in config.",
				},
				cli.BoolFlag{
					Name:  "no-recursive, quick",
					Usage: "Disable updating dependencies' dependencies. Only update things in glide.yaml.",
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "If there was a change in the repo or VCS switch to new one. Warning, changes will be lost.",
				},
				cli.BoolFlag{
					Name:  "update-vendored, u",
					Usage: "Update vendored packages (without local VCS repo). Warning, changes will be lost.",
				},
				cli.StringFlag{
					Name:  "file, f",
					Usage: "Save all of the discovered dependencies to a Glide YAML file.",
				},
			},
			Action: func(c *cli.Context) {
				cxt.Put("deleteOptIn", c.Bool("delete"))
				cxt.Put("forceUpdate", c.Bool("force"))
				cxt.Put("skipFlatten", c.Bool("no-recursive"))
				cxt.Put("deleteFlatten", c.Bool("delete-flatten"))
				cxt.Put("toPath", c.String("file"))
				cxt.Put("toStdout", false)
				if c.Bool("import") {
					cxt.Put("importGodeps", true)
					cxt.Put("importGPM", true)
					cxt.Put("importGb", true)
				}
				cxt.Put("updateVendoredDeps", c.Bool("update-vendored"))

				cxt.Put("packages", []string(c.Args()))
				setupHandler(c, "update", cxt, router)
			},
		},
		{
			Name:  "tree",
			Usage: "Tree prints the dependencies of this project as a tree.",
			Description: `This scans a project's source files and builds a tree
	representation of the import graph.

	It ignores testdata/ and directories that begin with . or _. Packages in
	vendor/ are only included if they are referenced by the main project or
	one of its dependencies.`,
			Action: func(c *cli.Context) {
				setupHandler(c, "tree", cxt, router)
			},
		},
		{
			Name:  "list",
			Usage: "List prints all dependencies that Glide could discover.",
			Description: `List scans your code and lists all of the packages that are used.

			It does not use the glide.yaml. Instead, it inspects the code to determine what packages are
			imported.

			Directories that begin with . or _ are ignored, as are testdata directories. Packages in
			vendor are only included if they are used by the project.
			`,
			Action: func(c *cli.Context) {
				setupHandler(c, "list", cxt, router)
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
	cxt.Put("debug", c.GlobalBool("debug"))
	cxt.Put("yaml", c.GlobalString("yaml"))
	cxt.Put("home", c.GlobalString("home"))
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
		Using("debug").From("cxt:debug").
		Does(cmd.VersionGuard, "v")

	reg.Route("@ready", "Prepare for glide commands.").
		Does(cmd.ReadyToGlide, "ready").Using("filename").From("cxt:yaml").
		Does(cmd.ParseYaml, "cfg").Using("filename").From("cxt:yaml").
		Does(cmd.EnsureCacheDir, "_").Using("home").From("cxt:home")

	reg.Route("get", "Install a pkg in vendor, and store the results in the glide.yaml").
		Includes("@startup").
		Includes("@ready").
		Does(cmd.GetAll, "goget").
		Using("filename").From("cxt:yaml").
		Using("packages").From("cxt:packages").
		Using("conf").From("cxt:cfg").
		Using("insecure").From("cxt:insecure").
		Using("home").From("cxt:home").
		Does(cmd.Flatten, "flatten").Using("conf").From("cxt:cfg").
		Using("packages").From("cxt:packages").
		Using("force").From("cxt:forceUpdate").
		Using("home").From("cxt:home").
		Does(cmd.WriteYaml, "out").
		Using("conf").From("cxt:cfg").
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
		Does(cmd.DeleteUnusedPackages, "deleted").
		Using("conf").From("cxt:cfg").
		Using("optIn").From("cxt:deleteOptIn").
		Does(cmd.VendoredSetup, "cfg").
		Using("conf").From("cxt:cfg").
		Using("update").From("cxt:updateVendoredDeps").
		Does(cmd.UpdateImports, "dependencies").
		Using("conf").From("cxt:cfg").
		Using("force").From("cxt:forceUpdate").
		Using("packages").From("cxt:packages").
		Using("home").From("cxt:home").
		Does(cmd.SetReference, "version").Using("conf").From("cxt:cfg").
		Does(cmd.Flatten, "flattened").Using("conf").From("cxt:cfg").
		Using("packages").From("cxt:packages").
		Using("force").From("cxt:forceUpdate").
		Using("skip").From("cxt:skipFlatten").
		Using("home").From("cxt:home").
		Does(cmd.VendoredCleanUp, "_").
		Using("conf").From("cxt:cfg").
		Using("update").From("cxt:updateVendoredDeps").
		Does(cmd.WriteYaml, "out").
		Using("conf").From("cxt:cfg").
		Using("filename").From("cxt:toPath").
		Using("toStdout").From("cxt:toStdout")

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
		Does(cmd.WriteYaml, "out").
		Using("conf").From("cxt:cfg").
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
		Does(cmd.WriteYaml, "out").Using("conf").From("cxt:cfg").
		Using("filename").From("cxt:toPath")

	reg.Route("import godep", "Read a Godeps.json file").
		Includes("@startup").
		Includes("@ready").
		Does(cmd.ParseGodepGodeps, "godeps").
		Does(cmd.AddDependencies, "addGodeps").
		Using("dependencies").From("cxt:godeps").
		Using("conf").From("cxt:cfg").
		// Does(cmd.UpdateReferences, "refs").Using("conf").From("cxt:cfg").
		Does(cmd.WriteYaml, "out").Using("conf").From("cxt:cfg").
		Using("filename").From("cxt:toPath")

	reg.Route("import gb", "Read a vendor/manifest file").
		Includes("@startup").
		Includes("@ready").
		Does(cmd.GbManifest, "manifest").
		Does(cmd.AddDependencies, "addGodeps").
		Using("dependencies").From("cxt:manifest").
		Using("conf").From("cxt:cfg").
		Does(cmd.WriteYaml, "out").Using("conf").From("cxt:cfg").
		Using("filename").From("cxt:toPath")

	reg.Route("guess", "Guess dependencies").
		Includes("@ready").
		Does(cmd.GuessDeps, "cfg").
		Does(cmd.WriteYaml, "out").
		Using("conf").From("cxt:cfg").
		Using("filename").From("cxt:toPath")

	reg.Route("create", "Initialize Glide").
		Includes("@startup").
		Does(cmd.InitGlide, "init").
		Using("filename").From("cxt:yaml").
		Using("project").From("cxt:project").WithDefault("main")

	reg.Route("name", "Print environment").
		Includes("@startup").
		Includes("@ready").
		Does(cmd.PrintName, "status").
		Using("conf").From("cxt:cfg")

	reg.Route("tree", "Print a dependency graph.").
		Includes("@startup").
		Does(cmd.Tree, "tree")
	reg.Route("list", "Print a dependency graph.").
		Includes("@startup").
		Does(cmd.ListDeps, "list")

	reg.Route("nv", "No Vendor").
		Includes("@startup").
		Does(cmd.NoVendor, "paths").
		Does(cmd.PathString, "out").Using("paths").From("cxt:paths")

	reg.Route("about", "Status").
		Includes("@startup").
		Does(cmd.About, "about")

	reg.Route("@plugin", "Try to send to a plugin.").
		Includes("@ready").
		Does(cmd.DropToShell, "plugin").
		Using("command").From("cxt:command")
}

func defaultGlideDir() string {
	c, err := user.Current()
	if err != nil {
		return ""
	}
	return filepath.Join(c.HomeDir, ".glide")
}
