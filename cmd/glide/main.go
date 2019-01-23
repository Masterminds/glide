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

	"github.com/Masterminds/glide"
	"github.com/Masterminds/glide/action"
	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"

	"github.com/codegangsta/cli"

	"os"
)

var version = "0.13.2-dev"

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
	app.Commands = glide.Commands()

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
