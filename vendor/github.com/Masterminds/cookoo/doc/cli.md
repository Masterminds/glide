# Writing a CLI

**Note:** There is now an official [CLI Tutorial](https://github.com/Masterminds/cookoo-cli-tutorial)
that explains by example how to build Cookoo CLIs.

The shortest Cookoo CLI you can write is this:

```go
package main

import (
	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/cookoo/cli"
)

func main() {
	cli.New(cookoo.Cookoo()).Run("help")
}
```

This, however, doesn't do much. It simply prints the generic help text
and exits.

## New-Style (Simple) CLI

The new-style CLI is great for building either a simple single-command
CLI or a more complex command with subcommands.

Here is a more representative CLI app:

```go
package main

import (
	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/cookoo/cli"
	"github.com/Masterminds/cookoo/fmt"

	"flag"
)

var Summary = "Example CLI"
var Description = "Full help text goes here."

func main() {
	reg, router, cxt := cookoo.Cookoo()

  // Create flags
	flags := flag.NewFlagSet("global", flag.PanicOnError)
	flags.Bool("h", false, "Show help")

	// A simple hello world route
	reg.Route("hello", "Show hello message").
		Does(fmt.Printf, "out").Using("format").WithDefault("Hello World\n")

  // This creates and executes a CLI app.
	cli.New(reg, router, cxt).Help(Summary, Description, flags).Run("hello")
}
```

Assuming the above has been compiled into `example`, here are the ways
the above could be called:

```
$ example     # Prints "Hello World"
$ example -h  # Prints help text
```

With a small change, we can re-build our app to support subcommands. We
can change the last line of the `main()` function to this:

```go
cli.New(reg, router, cxt).Help(Summary, Description, flags).RunSubcommand
```

Now we can call it in the following ways:

```
$ example        # Prints help text
$ example -h     # Prints help text
$ example help   # Prints help text
$ example hello  # Prints "Hello World"
```

This is all you need to begin writing new-style CLI apps.

### A Few Helpful Facts

- The context (`cxt`) contains `os.Args` and `runner.Args`, where the
  first is the OS args, and the second are the args remaining after
  flags are parsed.
- The flags are always available in the context as `globalFlags`.
- While you are **strongly** encouraged to use `Runner.Help()`, if you
  don't, very basic help text will be generated.
- If you want to use flags on your subcommands, you can parse those
  using the `cli.ParseArgs` command on your route. Pass it
  `runner.Args`.

Everything else in this document describes complex use cases.

## Old-style or Special Case CLIs

Here is a basic scaffold of a CLI application that uses subcommands.
Here, we're building the command `foo` that has the subcommand `bar`,
which can be invoked like this from tthe commandline: `foo bar`.

foo.go:
```go
package main

import(
  "github.com/Masterminds/cookoo"
  "github.com/Masterminds/cookoo/cli"
  "fmt"
  "os"
)

func main() {
	// Start a cookoo app.
	reg, router, cxt := cookoo.Cookoo();

	// Put the arguments into the context.
	cxt.Put("os.Args", os.Args)

	// Create help text
	reg.Route("help", "Show application-scoped help.").
		Does(cli.ShowHelp, "help").
			Using("show").WithDefault(true).
			Using("summary").WithDefault("This is the help text.")

	// Handle the "bar" subcommand
	reg.Route("bar", "Do something").
		Does(MyBarCommand, "bar")

	// This is the main runner. It proxies to subcommands.
	reg.Route("run", "Run the app.").
		// Shift off two args: the app name and then the subcommand.
		// The subcommand is then put in the context as "subcommand"
		Does(cli.ShiftArgs, "subcommand").
			Using("n").WithDefault(2).
		Does(cookoo.ForwardTo, "sub").
			Using("route").From("cxt:subcommand").WithDefault("help").
			Using("ignoreRoutes").WithDefault([]string{"subcommand"}).

	// This starts the app.	If a fatal error occurs, we
	// display the error.
	e := router.HandleRequest("run", cxt, true)
	if e != nil {
		fmt.Printf("Error: %s\n", e)
	}
}

// This is our command.
func MyBarCommand(cxt cookoo.Context, params *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fmt.Printf("OH HAI")
	return nil, nil
}
```

When we run `foo bar` (or `go run foo.go bar`), here's what happens:

1. main() is run. It creates a Cookoo app, defines the registry, and
   then runs `router.HandleRequest("run", ...)`
2. The "run" route is run. This reads the `os.Args` and sees that it
   should execute the subcommand `bar`.
3. The "bar" route is run, which executes it's one command:
   `MyBarCommand`.
4. `MyBarCommand` runs, printing "OH HAI" to stdout.

If you were to run `foo` or `foo help`, then the chain would execute
like this:

1. main() runs, and passes to `router.HandleRequest("run"...)
2. "run" will execute `ShiftArgs` and then `ForwardTo`, which will resolve the subcommand
   to "help" (which is the default target).
3. The "help" route will be run, which will print out simple help:

```
go run foo.go
SUMMARY

This is the help text.
```

## Using Flags (Advanced)

Go provides the `flag` package with many utilities for working with
command-line flags. We can use those in Cookoo.

Here's a snippet of code that could be worked into the previous example:

``` 
	reg.Route("add-user", "Add a new account.").
		Does(cli.ParseArgs, "args").
			Using("flagset").WithDefault(AddUserFlags()).
			Using("args").From("cxt:os.Args").
		Does(cli.ShowHelp, "help").
			Using("show").From("cxt:h").
			Using("summary").WithDefault("Add a new account.").
			Using("usage").WithDefault("tool add-user -a NAME -p PASSWORD").
			Using("flags").WithDefault(AccountAddFlags()).
		// ... do the rest...
```

With the example above, we can generate our flags like this:

```
func AddUserFlags() *flag.FlagSet {
	flags := flag.NewFlagSet("account", flag.PanicOnError)
	flags.Bool("h", false, "Print help text.")
	flags.String("a", nil, "Account name.")
	flags.String('p', nil, "User's new password.")

	return flags
}
```

The arguments are parsed by `cli.ParseArgs`. Each flag is then placed
directly into the context. (See how we access `-h` with `From(cxt:h)`
above?)

In addition to using flags for processing, `cli.ShowHelp` can take a
`*flag.FlagSet` and automatically generate help text.

While this example shows providing flags for subcommands, higher level
flags can be handled in much the same way. For example, we could use
them in our `run` route, too.


## Where to from here?

From this starting point, you should be able to assemble your own
routes, where each new route represents a subcommand.
