// Glide is a command line utility that manages Go project dependencies.
//
// Configuration of where to start is managed via a glide.yaml in the root of a
// project. The yaml
//
// A glide.yaml file looks like:
//
//		package: github.com/Masterminds/glide
//		imports:
//		- package: github.com/Masterminds/cookoo
//		- package: github.com/kylelemons/go-gypsy
//		  subpackages:
//		  - yaml
//
// Glide puts dependencies in a vendor directory. Go utilities require this to
// be in your GOPATH. Glide makes this easy.
//
// For more information use the `glide help` command or see https://glide.sh
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

var version = "0.13.0-dev"

const usage = `Vendor Package Management for your Go projects.

   Each project should have a 'glide.yaml' file in the project directory. Files
   look something like this:

       package: github.com/Masterminds/glide
       imports:
       - package: github.com/Masterminds/cookoo
         version: 1.1.0
       - package: github.com/kylelemons/go-gypsy
         subpackages:
         - yaml

   For more details on the 'glide.yaml' files see the documentation at
   https://glide.sh/docs/glide.yaml
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
			Name:  "debug",
			Usage: "Print debug verbose informational messages",
		},
		cli.StringFlag{
			Name:   "home",
			Value:  gpath.Home(),
			Usage:  "The location of Glide files",
			EnvVar: "GLIDE_HOME",
		},
		cli.StringFlag{
			Name:   "tmp",
			Value:  "",
			Usage:  "The temp directory to use. Defaults to systems temp",
			EnvVar: "GLIDE_TMP",
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

	// If there was an Error message exit non-zero.
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
			Action: func(c *cli.Context) error {
				action.Create(".", c.Bool("skip-import"), c.Bool("non-interactive"))
				return nil
			},
		},
		{
			Name:      "config-wizard",
			ShortName: "cw",
			Usage:     "Wizard that makes optional suggestions to improve config in a glide.yaml file.",
			Description: `Glide will analyze a projects glide.yaml file and the imported
		projects to find ways the glide.yaml file can potentially be improved. It
		will then interactively make suggestions that you can skip or accept.`,
			Action: func(c *cli.Context) error {
				action.ConfigWizard(".")
				return nil
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
   GPM, GOM, or GB Glide that configuration will be used to find the dependencies
   and versions to fetch. If those are not available the dependent packages will
   be fetched as either a version specified elsewhere or the latest version.

   When adding a new dependency Glide will perform an update to work out
   the versions for the dependencies of this dependency (transitive ones). This
   will generate an updated glide.lock file with specific locked versions to use.

   The '--strip-vendor' flag will remove any nested 'vendor' folders and
   'Godeps/_workspace' folders after an update (along with undoing any Godep
   import rewriting). Note, The Godeps specific functionality is deprecated and
   will be removed when most Godeps users have migrated to using the vendor
   folder.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "test",
					Usage: "Add test dependencies.",
				},
				cli.BoolFlag{
					Name:  "insecure",
					Usage: "Use http:// rather than https:// to retrieve packages.",
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
					Name:   "update-vendored, u",
					Usage:  "Update vendored packages (without local VCS repo). Warning, changes will be lost.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:   "cache",
					Usage:  "When downloading dependencies attempt to cache them.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:   "cache-gopath",
					Usage:  "When downloading dependencies attempt to put them in the GOPATH, too.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:   "use-gopath",
					Usage:  "Copy dependencies from the GOPATH if they exist there.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:  "resolve-current",
					Usage: "Resolve dependencies for only the current system rather than all build modes.",
				},
				cli.BoolFlag{
					Name:   "strip-vcs, s",
					Usage:  "Removes version control metadata (e.g, .git directory) from the vendor folder.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:  "strip-vendor, v",
					Usage: "Removes nested vendor and Godeps/_workspace directories.",
				},
				cli.BoolFlag{
					Name:  "non-interactive",
					Usage: "Disable interactive prompts.",
				},
				cli.BoolFlag{
					Name:  "skip-test",
					Usage: "Resolve dependencies in test files.",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Bool("delete") {
					msg.Warn("The --delete flag is deprecated. This now works by default.")
				}
				if c.Bool("update-vendored") {
					msg.Warn("The --update-vendored flag is deprecated. This now works by default.")
				}
				if c.String("file") != "" {
					msg.Warn("The --file flag is deprecated.")
				}
				if c.Bool("cache") {
					msg.Warn("The --cache flag is deprecated. This now works by default.")
				}
				if c.Bool("cache-gopath") {
					msg.Warn("The --cache-gopath flag is deprecated.")
				}
				if c.Bool("use-gopath") {
					msg.Warn("The --use-gopath flag is deprecated. Please see overrides.")
				}
				if c.Bool("strip-vcs") {
					msg.Warn("The --strip-vcs flag is deprecated. This now works by default.")
				}

				if len(c.Args()) < 1 {
					fmt.Println("Oops! Package name is required.")
					os.Exit(1)
				}

				if c.Bool("resolve-current") {
					util.ResolveCurrent = true
					msg.Warn("Only resolving dependencies for the current OS/Arch.")
				}

				inst := repo.NewInstaller()
				inst.Force = c.Bool("force")
				inst.ResolveAllFiles = c.Bool("all-dependencies")
				inst.ResolveTest = !c.Bool("skip-test")
				packages := []string(c.Args())
				insecure := c.Bool("insecure")
				action.Get(packages, inst, insecure, c.Bool("no-recursive"), c.Bool("strip-vendor"), c.Bool("non-interactive"), c.Bool("test"))
				return nil
			},
		},
		{
			Name:      "remove",
			ShortName: "rm",
			Usage:     "Remove a package from the glide.yaml file, and regenerate the lock file.",
			Description: `This takes one or more package names, and removes references from the glide.yaml file.
   This will rebuild the glide lock file re-resolving the depencies.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "delete,d",
					Usage: "Also delete from vendor/ any packages that are no longer used.",
				},
			},
			Action: func(c *cli.Context) error {
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
				return nil
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
					Action: func(c *cli.Context) error {
						action.ImportGodep(c.String("file"))
						return nil
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
					Action: func(c *cli.Context) error {
						action.ImportGPM(c.String("file"))
						return nil
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
					Action: func(c *cli.Context) error {
						action.ImportGB(c.String("file"))
						return nil
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
					Action: func(c *cli.Context) error {
						action.ImportGom(c.String("file"))
						return nil
					},
				},
			},
		},
		{
			Name:        "name",
			Usage:       "Print the name of this project.",
			Description: `Read the glide.yaml file and print the name given on the 'package' line.`,
			Action: func(c *cli.Context) error {
				action.Name()
				return nil
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
			Action: func(c *cli.Context) error {
				action.NoVendor(c.String("dir"), true, !c.Bool("no-subdir"))
				return nil
			},
		},
		{
			Name:  "rebuild",
			Usage: "Rebuild ('go build') the dependencies",
			Description: `(Deprecated) This rebuilds the packages' '.a' files. On some systems
	this can improve performance on subsequent 'go run' and 'go build' calls.`,
			Action: func(c *cli.Context) error {
				action.Rebuild()
				return nil
			},
		},
		{
			Name:      "install",
			ShortName: "i",
			Usage:     "Install a project's dependencies",
			Description: `This uses the native VCS of each package to install
   the appropriate version. There are two ways a project's dependencies can
   be installed. When there is a glide.yaml file defining the dependencies but
   no lock file (glide.lock) the dependencies are installed using the "update"
   command and a glide.lock file is generated pinning all dependencies. If a
   glide.lock file is already present the dependencies are installed or updated
   from the lock file.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:   "delete",
					Usage:  "Delete vendor packages not specified in config.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "If there was a change in the repo or VCS switch to new one. Warning: changes will be lost.",
				},
				cli.BoolFlag{
					Name:   "update-vendored, u",
					Usage:  "Update vendored packages (without local VCS repo). Warning: this may destroy local modifications to vendor/.",
					Hidden: true,
				},
				cli.StringFlag{
					Name:   "file, f",
					Usage:  "Save all of the discovered dependencies to a Glide YAML file. (DEPRECATED: This has no impact.)",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:   "cache",
					Usage:  "When downloading dependencies attempt to cache them.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:   "cache-gopath",
					Usage:  "When downloading dependencies attempt to put them in the GOPATH, too.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:   "use-gopath",
					Usage:  "Copy dependencies from the GOPATH if they exist there.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:   "strip-vcs, s",
					Usage:  "Removes version control metadata (e.g, .git directory) from the vendor folder.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:  "strip-vendor, v",
					Usage: "Removes nested vendor and Godeps/_workspace directories.",
				},
				cli.BoolFlag{
					Name:  "skip-test",
					Usage: "Resolve dependencies in test files.",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Bool("delete") {
					msg.Warn("The --delete flag is deprecated. This now works by default.")
				}
				if c.Bool("update-vendored") {
					msg.Warn("The --update-vendored flag is deprecated. This now works by default.")
				}
				if c.String("file") != "" {
					msg.Warn("The --flag flag is deprecated.")
				}
				if c.Bool("cache") {
					msg.Warn("The --cache flag is deprecated. This now works by default.")
				}
				if c.Bool("cache-gopath") {
					msg.Warn("The --cache-gopath flag is deprecated.")
				}
				if c.Bool("use-gopath") {
					msg.Warn("The --use-gopath flag is deprecated. Please see overrides.")
				}
				if c.Bool("strip-vcs") {
					msg.Warn("The --strip-vcs flag is deprecated. This now works by default.")
				}

				installer := repo.NewInstaller()
				installer.Force = c.Bool("force")
				installer.Home = c.GlobalString("home")
				installer.ResolveTest = !c.Bool("skip-test")

				action.Install(installer, c.Bool("strip-vendor"))
				return nil
			},
		},
		{
			Name:      "update",
			ShortName: "up",
			Usage:     "Update a project's dependencies",
			Description: `This updates the dependencies by scanning the codebase
   to determine the needed dependencies and fetching them following the rules
   in the glide.yaml file. When no rules exist the tip of the default branch
   is used. For more details see https://glide.sh/docs/glide.yaml

   If a dependency has a glide.yaml file, update will read that file and
   use the information contained there. Those dependencies are maintained in
   the top level 'vendor/' directory. 'vendor/foo/bar' will have its
   dependencies stored in 'vendor/'. This behavior can be disabled with
   '--no-recursive'. When this behavior is skipped a glide.lock file is not
   generated because the full dependency tree cannot be known.

   Glide will also import Godep, GB, GOM, and GPM files as it finds them in dependencies.
   It will create a glide.yaml file from the Godeps data, and then update. This
   has no effect if '--no-recursive' is set.

   The '--strip-vendor' flag will remove any nested 'vendor' folders and
   'Godeps/_workspace' folders after an update (along with undoing any Godep
   import rewriting). Note, the Godeps specific functionality is deprecated and
   will be removed when most Godeps users have migrated to using the vendor
   folder.`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:   "delete",
					Usage:  "Delete vendor packages not specified in config.",
					Hidden: true,
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
					Name:   "update-vendored, u",
					Usage:  "Update vendored packages (without local VCS repo). Warning, changes will be lost.",
					Hidden: true,
				},
				cli.StringFlag{
					Name:   "file, f",
					Usage:  "Save all of the discovered dependencies to a Glide YAML file.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:   "cache",
					Usage:  "When downloading dependencies attempt to cache them.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:   "cache-gopath",
					Usage:  "When downloading dependencies attempt to put them in the GOPATH, too.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:   "use-gopath",
					Usage:  "Copy dependencies from the GOPATH if they exist there.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:  "resolve-current",
					Usage: "Resolve dependencies for only the current system rather than all build modes.",
				},
				cli.BoolFlag{
					Name:   "strip-vcs, s",
					Usage:  "Removes version control metadata (e.g, .git directory) from the vendor folder.",
					Hidden: true,
				},
				cli.BoolFlag{
					Name:  "strip-vendor, v",
					Usage: "Removes nested vendor and Godeps/_workspace directories.",
				},
				cli.BoolFlag{
					Name:  "skip-test",
					Usage: "Resolve dependencies in test files.",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Bool("delete") {
					msg.Warn("The --delete flag is deprecated. This now works by default.")
				}
				if c.Bool("update-vendored") {
					msg.Warn("The --update-vendored flag is deprecated. This now works by default.")
				}
				if c.String("file") != "" {
					msg.Warn("The --flag flag is deprecated.")
				}
				if c.Bool("cache") {
					msg.Warn("The --cache flag is deprecated. This now works by default.")
				}
				if c.Bool("cache-gopath") {
					msg.Warn("The --cache-gopath flag is deprecated.")
				}
				if c.Bool("use-gopath") {
					msg.Warn("The --use-gopath flag is deprecated. Please see overrides.")
				}
				if c.Bool("strip-vcs") {
					msg.Warn("The --strip-vcs flag is deprecated. This now works by default.")
				}

				if c.Bool("resolve-current") {
					util.ResolveCurrent = true
					msg.Warn("Only resolving dependencies for the current OS/Arch")
				}

				installer := repo.NewInstaller()
				installer.Force = c.Bool("force")
				installer.ResolveAllFiles = c.Bool("all-dependencies")
				installer.Home = c.GlobalString("home")
				installer.ResolveTest = !c.Bool("skip-test")

				action.Update(installer, c.Bool("no-recursive"), c.Bool("strip-vendor"))

				return nil
			},
		},
		{
			Name:  "tree",
			Usage: "(Deprecated) Tree prints the dependencies of this project as a tree.",
			Description: `This scans a project's source files and builds a tree
   representation of the import graph.

   It ignores testdata/ and directories that begin with . or _. Packages in
   vendor/ are only included if they are referenced by the main project or
   one of its dependencies.

   Note, for large projects this can display a large list tens of thousands of
   lines long.`,
			Action: func(c *cli.Context) error {
				action.Tree(".", false)
				return nil
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
			Action: func(c *cli.Context) error {
				action.List(".", true, c.String("output"))
				return nil
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
			Action: func(c *cli.Context) error {
				if c.IsSet("format") {
					action.Info(c.String("format"))
				} else {
					cli.ShowCommandHelp(c, c.Command.Name)
				}
				return nil
			},
		},
		{
			Name:      "cache-clear",
			ShortName: "cc",
			Usage:     "Clears the Glide cache.",
			Action: func(c *cli.Context) error {
				action.CacheClear()
				return nil
			},
		},
		{
			Name:  "about",
			Usage: "Learn about Glide",
			Action: func(c *cli.Context) error {
				action.About()
				return nil
			},
		},
		{
			Name:  "mirror",
			Usage: "Manage mirrors",
			Description: `Mirrors provide the ability to replace a repo location with
   another location that's a mirror of the original. This is useful when you want
   to have a cache for your continuous integration (CI) system or if you want to
   work on a dependency in a local location.

   The mirrors are stored in a mirrors.yaml file in your GLIDE_HOME.

   The three commands to manage mirrors are 'list', 'set', and 'remove'.

   Use 'set' in the form:

       glide mirror set [original] [replacement]

   or

       glide mirror set [original] [replacement] --vcs [type]

   for example,

       glide mirror set https://github.com/example/foo https://git.example.com/example/foo.git

       glide mirror set https://github.com/example/foo file:///path/to/local/repo --vcs git

   Use 'remove' in the form:

       glide mirror remove [original]

   for example,

       glide mirror remove https://github.com/example/foo`,
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List the current mirrors",
					Action: func(c *cli.Context) error {
						return action.MirrorsList()
					},
				},
				{
					Name:  "set",
					Usage: "Set a mirror. This overwrites an existing entry if one exists",
					Description: `Use 'set' in the form:

       glide mirror set [original] [replacement]

   or

       glide mirror set [original] [replacement] --vcs [type]

   for example,

       glide mirror set https://github.com/example/foo https://git.example.com/example/foo.git

       glide mirror set https://github.com/example/foo file:///path/to/local/repo --vcs git`,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "vcs",
							Usage: "The VCS type to use. Autodiscovery is attempted when not supplied. Can be one of git, svn, bzr, or hg",
						},
					},
					Action: func(c *cli.Context) error {
						return action.MirrorsSet(c.Args().Get(0), c.Args().Get(1), c.String("vcs"))
					},
				},
				{
					Name:      "remove",
					ShortName: "rm",
					Usage:     "Remove a mirror",
					Description: `Use 'remove' in the form:

       glide mirror remove [original]

   for example,

       glide mirror remove https://github.com/example/foo`,
					Action: func(c *cli.Context) error {
						return action.MirrorsRemove(c.Args().Get(0))
					},
				},
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
	action.NoColor(c.Bool("no-color"))
	action.Quiet(c.Bool("quiet"))
	action.Init(c.String("yaml"), c.String("home"))
	action.EnsureGoVendor()
	gpath.Tmp = c.String("tmp")
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
