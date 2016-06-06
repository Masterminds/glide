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

	"github.com/Masterminds/glide/action"
	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
	"github.com/Masterminds/glide/repo"
	"github.com/Masterminds/glide/util"

	"github.com/codegangsta/cli"

	"fmt"
	"os"
)

var version = "0.11.0-dev"

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

NOTE: As of Glide 0.5, the commands 'into', 'gopath', 'status', and 'env'
no longer exist.
`

// VendorDir default vendor directory name
var VendorDir = "vendor"

func main() {
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
			Name:  "verbose",
			Usage: "Print detailed informational messages",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Print debug verbose informational messages",
		},
		cli.StringFlag{
			Name:   "home",
			Value:  gpath.Home(),
			Usage:  "The location of Glide files",
			EnvVar: "GLIDE_HOME",
		},
		cli.BoolFlag{
			Name:  "no-color",
			Usage: "Turn off colored output for log messages",
		},
	}
	app.CommandNotFound = func(c *cli.Context, command string) {
		// TODO: Set some useful env vars.
		action.Plugin(command, os.Args)
	}
	app.Before = startup
	app.After = shutdown
	app.Commands = commands()

	// Detect errors from the Before and After calls and exit on them.
	if err := app.Run(os.Args); err != nil {
		msg.Err(err.Error())
		os.Exit(1)
	}

	// If there was a Error message exit non-zero.
	if msg.HasErrored() {
		m := msg.Color(msg.Red, "An Error has occurred")
		msg.Msg(m)
		os.Exit(2)
	}
}

func commands() []cli.Command {
	return []cli.Command{
		{
			Name:      "create",
			ShortName: "init",
			Usage:     "Initialize a new project, creating a glide.yaml file",
			Description: `This command starts from a project without Glide and
   sets it up. It generates a glide.yaml file, parsing your codebase to guess
   the dependencies to include. Once this step is done you may edit the
   glide.yaml file to update imported dependency properties such as the version
   or version range to include.

   To fetch the dependencies you may run 'glide install'.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "skip-import",
					Usage: "When initializing skip importing from other package managers.",
				},
				cli.BoolFlag{
					Name:  "non-interactive",
					Usage: "Disable interactive prompts.",
				},
			},
			Action: func(c *cli.Context) {
				action.Create(".", c.Bool("skip-import"), c.Bool("non-interactive"))
			},
		},
		{
			Name:      "config-wizard",
			ShortName: "cw",
			Usage:     "Wizard that makes optional suggestions to improve config in a glide.yaml file.",
			Description: `Glide will analyze a projects glide.yaml file and the imported
		projects to find ways the glide.yaml file can potentially be improved. It
		will then interactively make suggestions that you can skip or accept.`,
			Action: func(c *cli.Context) {
				action.ConfigWizard(".")
			},
		},
		{
			Name:  "get",
			Usage: "Install one or more packages into `vendor/` and add dependency to glide.yaml.",
			Description: `Gets one or more package (like 'go get') and then adds that file
   to the glide.yaml file. Multiple package names can be specified on one line.

   	$ glide get github.com/Masterminds/cookoo/web

   The above will install the project github.com/Masterminds/cookoo and add
   the subpackage 'web'.

   If a fetched dependency has a glide.yaml file, configuration from Godep,
   GPM, or GB Glide that configuration will be used to find the dependencies
   and versions to fetch. If those are not available the dependent packages will
   be fetched as either a version specified elsewhere or the latest version.

   When adding a new dependency Glide will perform an update to work out the
   the versions to use from the dependency tree. This will generate an updated
   glide.lock file with specific locked versions to use.

   If you are storing the outside dependencies in your version control system
   (VCS), also known as vendoring, there are a few flags that may be useful.
   The '--update-vendored' flag will cause Glide to update packages when VCS
   information is unavailable. This can be used with the '--strip-vcs' flag which
   will strip VCS data found in the vendor directory. This is useful for
   removing VCS data from transitive dependencies and initial setups. The
   '--strip-vendor' flag will remove any nested 'vendor' folders and
   'Godeps/_workspace' folders after an update (along with undoing any Godep
   import rewriting). Note, The Godeps specific functionality is deprecated and
   will be removed when most Godeps users have migrated to using the vendor
   folder.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "insecure",
					Usage: "Use http:// rather than https:// to retrieve pacakges.",
				},
				cli.BoolFlag{
					Name:  "no-recursive, quick",
					Usage: "Disable updating dependencies' dependencies.",
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "If there was a change in the repo or VCS switch to new one. Warning, changes will be lost.",
				},
				cli.BoolFlag{
					Name:  "all-dependencies",
					Usage: "This will resolve all dependencies for all packages, not just those directly used.",
				},
				cli.BoolFlag{
					Name:  "update-vendored, u",
					Usage: "Update vendored packages (without local VCS repo). Warning, changes will be lost.",
				},
				cli.BoolFlag{
					Name:  "cache",
					Usage: "When downloading dependencies attempt to cache them.",
				},
				cli.BoolFlag{
					Name:  "cache-gopath",
					Usage: "When downloading dependencies attempt to put them in the GOPATH, too.",
				},
				cli.BoolFlag{
					Name:  "use-gopath",
					Usage: "Copy dependencies from the GOPATH if they exist there.",
				},
				cli.BoolFlag{
					Name:  "resolve-current",
					Usage: "Resolve dependencies for only the current system rather than all build modes.",
				},
				cli.BoolFlag{
					Name:  "strip-vcs, s",
					Usage: "Removes version control metadata (e.g, .git directory) from the vendor folder.",
				},
				cli.BoolFlag{
					Name:  "strip-vendor, v",
					Usage: "Removes nested vendor and Godeps/_workspace directories. Requires --strip-vcs.",
				},
				cli.BoolFlag{
					Name:  "non-interactive",
					Usage: "Disable interactive prompts.",
				},
			},
			Action: func(c *cli.Context) {
				if c.Bool("strip-vendor") && !c.Bool("strip-vcs") {
					msg.Die("--strip-vendor cannot be used without --strip-vcs")
				}

				if len(c.Args()) < 1 {
					fmt.Println("Oops! Package name is required.")
					os.Exit(1)
				}

				if c.Bool("resolve-current") {
					util.ResolveCurrent = true
					msg.Warn("Only resolving dependencies for the current OS/Arch")
				}

				inst := repo.NewInstaller()
				inst.Force = c.Bool("force")
				inst.UseCache = c.Bool("cache")
				inst.UseGopath = c.Bool("use-gopath")
				inst.UseCacheGopath = c.Bool("cache-gopath")
				inst.UpdateVendored = c.Bool("update-vendored")
				inst.ResolveAllFiles = c.Bool("all-dependencies")
				packages := []string(c.Args())
				insecure := c.Bool("insecure")
				action.Get(packages, inst, insecure, c.Bool("no-recursive"), c.Bool("strip-vcs"), c.Bool("strip-vendor"), c.Bool("non-interactive"))
			},
		},
		{
			Name:      "remove",
			ShortName: "rm",
			Usage:     "Remove a package from the glide.yaml file, and regenerate the lock file.",
			Description: `This takes one or more package names, and removes references from the glide.yaml file.
   This will rebuild the glide lock file with the following constraints:

   - Dependencies are re-negotiated. Any that are no longer used are left out of the lock.
   - Minor version re-nogotiation is performed on remaining dependencies.
   - No updates are peformed. You may want to run 'glide up' to accomplish that.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "delete,d",
					Usage: "Also delete from vendor/ any packages that are no longer used.",
				},
			},
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					fmt.Println("Oops! At least one package name is required.")
					os.Exit(1)
				}

				if c.Bool("delete") {
					// FIXME: Implement this in the installer.
					fmt.Println("Delete is not currently implemented.")
				}
				inst := repo.NewInstaller()
				inst.Force = c.Bool("force")
				packages := []string(c.Args())
				action.Remove(packages, inst)
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
						action.ImportGodep(c.String("file"))
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
						action.ImportGPM(c.String("file"))
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
						action.ImportGB(c.String("file"))
					},
				},
				{
					Name:  "gom",
					Usage: "Import Gomfile and display the would-be yaml file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "file, f",
							Usage: "Save all of the discovered dependencies to a Glide YAML file.",
						},
					},
					Action: func(c *cli.Context) {
						action.ImportGom(c.String("file"))
					},
				},
			},
		},
		{
			Name:        "name",
			Usage:       "Print the name of this project.",
			Description: `Read the glide.yaml file and print the name given on the 'package' line.`,
			Action: func(c *cli.Context) {
				action.Name()
			},
		},
		{
			Name:      "novendor",
			ShortName: "nv",
			Usage:     "List all non-vendor paths in a directory.",
			Description: `Given a directory, list all the relevant Go paths that are not vendored.

Example:
   $ go test $(glide novendor)`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dir,d",
					Usage: "Specify a directory to run novendor against.",
					Value: ".",
				},
				cli.BoolFlag{
					Name:  "no-subdir,x",
					Usage: "Specify this to prevent nv from append '/...' to all directories.",
				},
			},
			Action: func(c *cli.Context) {
				action.NoVendor(c.String("dir"), true, !c.Bool("no-subdir"))
			},
		},
		{
			Name:  "rebuild",
			Usage: "Rebuild ('go build') the dependencies",
			Description: `This rebuilds the packages' '.a' files. On some systems
	this can improve performance on subsequent 'go run' and 'go build' calls.`,
			Action: func(c *cli.Context) {
				action.Rebuild()
			},
		},
		{
			Name:      "install",
			ShortName: "i",
			Usage:     "Install a project's dependencies",
			Description: `This uses the native VCS of each packages to install
   the appropriate version. There are two ways a projects dependencies can
   be installed. When there is a glide.yaml file defining the dependencies but
   no lock file (glide.lock) the dependencies are installed using the "update"
   command and a glide.lock file is generated pinning all dependencies. If a
   glide.lock file is already present the dependencies are installed or updated
   from the lock file.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "delete",
					Usage: "Delete vendor packages not specified in config.",
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "If there was a change in the repo or VCS switch to new one. Warning: changes will be lost.",
				},
				cli.BoolFlag{
					Name:  "update-vendored, u",
					Usage: "Update vendored packages (without local VCS repo). Warning: this may destroy local modifications to vendor/.",
				},
				cli.StringFlag{
					Name:  "file, f",
					Usage: "Save all of the discovered dependencies to a Glide YAML file. (DEPRECATED: This has no impact.)",
				},
				cli.BoolFlag{
					Name:  "cache",
					Usage: "When downloading dependencies attempt to cache them.",
				},
				cli.BoolFlag{
					Name:  "cache-gopath",
					Usage: "When downloading dependencies attempt to put them in the GOPATH, too.",
				},
				cli.BoolFlag{
					Name:  "use-gopath",
					Usage: "Copy dependencies from the GOPATH if they exist there.",
				},
				cli.BoolFlag{
					Name:  "strip-vcs, s",
					Usage: "Removes version control metadata (e.g, .git directory) from the vendor folder.",
				},
				cli.BoolFlag{
					Name:  "strip-vendor, v",
					Usage: "Removes nested vendor and Godeps/_workspace directories. Requires --strip-vcs.",
				},
			},
			Action: func(c *cli.Context) {
				if c.Bool("strip-vendor") && !c.Bool("strip-vcs") {
					msg.Die("--strip-vendor cannot be used without --strip-vcs")
				}

				installer := repo.NewInstaller()
				installer.Force = c.Bool("force")
				installer.UseCache = c.Bool("cache")
				installer.UseGopath = c.Bool("use-gopath")
				installer.UseCacheGopath = c.Bool("cache-gopath")
				installer.UpdateVendored = c.Bool("update-vendored")
				installer.Home = c.GlobalString("home")
				installer.DeleteUnused = c.Bool("delete")

				action.Install(installer, c.Bool("strip-vcs"), c.Bool("strip-vendor"))
			},
		},
		{
			Name:      "update",
			ShortName: "up",
			Usage:     "Update a project's dependencies",
			Description: `This uses the native VCS of each package to try to
   pull the most applicable updates. Packages with fixed refs (Versions or
   tags) will not be updated. Packages with no ref or with a branch ref will
   be updated as expected.

   If a dependency has a glide.yaml file, update will read that file and
   update those dependencies accordingly. Those dependencies are maintained in
   a the top level 'vendor/' directory. 'vendor/foo/bar' will have its
   dependencies stored in 'vendor/'. This behavior can be disabled with
   '--no-recursive'. When this behavior is skipped a glide.lock file is not
   generated because the full dependency tree cannot be known.

   Glide will also import Godep, GB, and GPM files as it finds them in dependencies.
   It will create a glide.yaml file from the Godeps data, and then update. This
   has no effect if '--no-recursive' is set.

   If you are storing the outside dependencies in your version control system
   (VCS), also known as vendoring, there are a few flags that may be useful.
   The '--update-vendored' flag will cause Glide to update packages when VCS
   information is unavailable. This can be used with the '--strip-vcs' flag which
   will strip VCS data found in the vendor directory. This is useful for
   removing VCS data from transitive dependencies and initial setups. The
   '--strip-vendor' flag will remove any nested 'vendor' folders and
   'Godeps/_workspace' folders after an update (along with undoing any Godep
   import rewriting). Note, The Godeps specific functionality is deprecated and
   will be removed when most Godeps users have migrated to using the vendor
   folder.

   Note, Glide detects vendored dependencies. With the '--update-vendored' flag
   Glide will update vendored dependencies leaving them in a vendored state.
   Tertiary dependencies will not be vendored automatically unless the
   '--strip-vcs' flag is used along with it.

   By default, packages that are discovered are considered transient, and are
   not stored in the glide.yaml file. The --file=NAME.yaml flag allows you
   to save the discovered dependencies to a YAML file.`,
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
					Name:  "all-dependencies",
					Usage: "This will resolve all dependencies for all packages, not just those directly used.",
				},
				cli.BoolFlag{
					Name:  "update-vendored, u",
					Usage: "Update vendored packages (without local VCS repo). Warning, changes will be lost.",
				},
				cli.StringFlag{
					Name:  "file, f",
					Usage: "Save all of the discovered dependencies to a Glide YAML file.",
				},
				cli.BoolFlag{
					Name:  "cache",
					Usage: "When downloading dependencies attempt to cache them.",
				},
				cli.BoolFlag{
					Name:  "cache-gopath",
					Usage: "When downloading dependencies attempt to put them in the GOPATH, too.",
				},
				cli.BoolFlag{
					Name:  "use-gopath",
					Usage: "Copy dependencies from the GOPATH if they exist there.",
				},
				cli.BoolFlag{
					Name:  "resolve-current",
					Usage: "Resolve dependencies for only the current system rather than all build modes.",
				},
				cli.BoolFlag{
					Name:  "strip-vcs, s",
					Usage: "Removes version control metadata (e.g, .git directory) from the vendor folder.",
				},
				cli.BoolFlag{
					Name:  "strip-vendor, v",
					Usage: "Removes nested vendor and Godeps/_workspace directories. Requires --strip-vcs.",
				},
			},
			Action: func(c *cli.Context) {
				if c.Bool("strip-vendor") && !c.Bool("strip-vcs") {
					msg.Die("--strip-vendor cannot be used without --strip-vcs")
				}

				if c.Bool("resolve-current") {
					util.ResolveCurrent = true
					msg.Warn("Only resolving dependencies for the current OS/Arch")
				}

				installer := repo.NewInstaller()
				installer.Force = c.Bool("force")
				installer.UseCache = c.Bool("cache")
				installer.UseGopath = c.Bool("use-gopath")
				installer.UseCacheGopath = c.Bool("cache-gopath")
				installer.UpdateVendored = c.Bool("update-vendored")
				installer.ResolveAllFiles = c.Bool("all-dependencies")
				installer.Home = c.GlobalString("home")
				installer.DeleteUnused = c.Bool("delete")

				action.Update(installer, c.Bool("no-recursive"), c.Bool("strip-vcs"), c.Bool("strip-vendor"))
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
				action.Tree(".", false)
			},
		},
		{
			Name:  "list",
			Usage: "List prints all dependencies that the present code references.",
			Description: `List scans your code and lists all of the packages that are used.

   It does not use the glide.yaml. Instead, it inspects the code to determine what packages are
   imported.

   Directories that begin with . or _ are ignored, as are testdata directories. Packages in
   vendor are only included if they are used by the project.`,
			Action: func(c *cli.Context) {
				action.List(".", true, c.String("output"))
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output, o",
					Usage: "Output format. One of: json|json-pretty|text",
					Value: "text",
				},
			},
		},
		{
			Name:  "info",
			Usage: "Info prints information about this project",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "format, f",
					Usage: `Format of the information wanted (required).`,
				},
			},
			Description: `A format containing the text with replacement variables
   has to be passed in. Those variables are:

       %n - name
       %d - description
       %h - homepage
       %l - license

   For example, given a project with the following glide.yaml:

       package: foo
       homepage: https://example.com
       license: MIT
       description: Some example description

   Then running the following commands:

       glide info -f %n
          prints 'foo'

       glide info -f "License: %l"
          prints 'License: MIT'

       glide info -f "%n - %d - %h - %l"
          prints 'foo - Some example description - https://example.com - MIT'`,
			Action: func(c *cli.Context) {
				if c.IsSet("format") {
					action.Info(c.String("format"))
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
				}
			},
		},
		{
			Name:      "cache-clear",
			ShortName: "cc",
			Usage:     "Clears the Glide cache.",
			Action: func(c *cli.Context) {
				action.CacheClear()
			},
		},
		{
			Name:  "about",
			Usage: "Learn about Glide",
			Action: func(c *cli.Context) {
				action.About()
			},
		},
	}
}

// startup sets up the base environment.
//
// It does not assume the presence of a Glide.yaml file or vendor/ directory,
// so it can be used by any Glide command.
func startup(c *cli.Context) error {
	action.Debug(c.Bool("debug"))
	action.Verbose(c.Bool("verbose"))
	action.NoColor(c.Bool("no-color"))
	action.Quiet(c.Bool("quiet"))
	action.Init(c.String("yaml"), c.String("home"))
	action.EnsureGoVendor()
	return nil
}

func shutdown(c *cli.Context) error {
	cache.SystemUnlock()
	return nil
}

// Get the path to the glide.yaml file.
//
// This returns the name of the path, even if the file does not exist. The value
// may be set by the user, or it may be the default.
func glidefile(c *cli.Context) string {
	path := c.String("file")
	if path == "" {
		// For now, we construct a basic assumption. In the future, we could
		// traverse backward to see if a glide.yaml exists in a parent.
		path = "./glide.yaml"
	}
	a, err := filepath.Abs(path)
	if err != nil {
		// Underlying fs didn't provide working dir.
		return path
	}
	return a
}
