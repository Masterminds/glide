package cli

import (
	"flag"
	"fmt"
	"github.com/Masterminds/cookoo"
	"io"
	"os"
	"strings"
	//"log"
)

// Parse arguments for a "subcommand"
//
// The cookoo.cli.RequestResolver allows you to specify global level flags. This command
// allows you to augment those with subcommand flags. Example:
//
// 		$ myprog -foo=yes subcommand -bar=no
//
// In the above example, `-foo` is a global flag (set before the subcommand), while
// `-bar` is a local flag. It is specific to `subcommand`. This command lets you parse
// an arguments list given a pointer to a `flag.FlagSet`.
//
// Like the cookoo.cli.RequestResolver, this will place the parsed params directly into the
// context. For this reason, you ought not use the same flag names at both global and local
// flag levels. (The local will overwrite the global.)
//
// Params:
//
// 	- args: (required) A slice of arguments. Typically, this is `cxt:args` as set by
// 		cookoo.cli.RequestResolver.
// 	- flagset: (required) A set if flags (see flag.FlagSet) to parse.
//
// A slice of all non-flag arguments remaining after the parse are returned into the context.
//
// For example, if ['-foo', 'bar', 'some', 'other', 'data'] is passed in, '-foo' and 'bar' will
// be parsed out, while ['some', 'other', 'data'] will be returned into the context. (Assuming, of
// course, that the flag definition for -foo exists, and is a type that accepts a value).
//
// Thus, you will have `cxt:foo` available (with value `bar`) and everything else will be available
// in the slice under this command's context entry.
func ParseArgs(cxt cookoo.Context, params *cookoo.Params) (interface{}, cookoo.Interrupt) {
	params.Requires("args", "flagset")
	flagset := params.Get("flagset", nil).(*flag.FlagSet)
	args := params.Get("args", nil).([]string)

	// If this is true, we shift the args first.
	if params.Get("subcommand", false).(bool) {
		args = args[1:]
	}


	flagset.Parse(args)
	addFlagsToContext(flagset, cxt)
	return flagset.Args(), nil

}

// Show help.
// This command is useful for placing at the front of a CLI "subcommand" to have it output
// help information. It will only trigger when "show" is set to true, so another command
// can, for example, check for a "-h" or "-help" flag and set "show" based on that.
//
// Params:
// 	- show (bool): If `true`, show help.
// 	- summary (string): A one-line summary of the command.
// 	- description (string): A short description of what the command does.
// 	- usage (string): usage information.
// 	- flags (FlagSet): Flags that are supported. The FlagSet will be converted to help text.
// 	- writer (Writer): The location that this will write to. Default is os.Stdout
// 	- subcommands ([]string): A list of subcommands. This will be formatted as help text.
func ShowHelp(cxt cookoo.Context, params *cookoo.Params) (interface{}, cookoo.Interrupt) {
	showHelp := false
	showHelpO := params.Get("show", false)
	switch showHelpO.(type) {
	case string:
		showHelp = strings.ToLower(showHelpO.(string)) == "true"
	case bool:
		showHelp = showHelpO.(bool)
	}

	writer := params.Get("writer", os.Stdout).(io.Writer)

	pmap := params.AsMap()

	// Last resort: If no summary, pull it from the route description.
	if summary, ok := pmap["summary"]; !ok || len(summary.(string)) == 0 {
		pmap["summary"] = cxt.Get("route.Description", "").(string)
	}

	sections := []string{"summary", "description", "usage"}
	if _, ok := params.Has("subcommands"); ok {
		sections = append(sections, "subcommands")
	}

	if showHelp {
		displayHelp(sections, pmap, writer)
		return true, new(cookoo.Stop)
	}

	return false, nil
}

func displayHelp(keys []string, params map[string]interface{}, out io.Writer) {
	for i := range keys {
		key := keys[i]
		if msg, ok := params[key]; ok && len(msg.(string)) > 0 {
			spacer := strings.Repeat("=", len(key))
			fmt.Fprintf(out, "\n%s\n%s\n\n%s\n", strings.ToUpper(key), spacer, msg)
		}
	}
	fmt.Fprintf(out, "\n")

	// Handle the flags, if set.
	args, ok := params["flags"]
	if ok {
		fmt.Fprintf(out, "FLAGS\n=====\n")
		// Name: Description (default)
		filter := "\t-%s: %s (Default: '%s')\n"
		args.(*flag.FlagSet).VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(out, filter, f.Name, f.Usage, f.DefValue)
		})
	}
}

// Run a subcommand.
//
// Params:
// 	- args: a string[] of arguments, like you get from os.Args. This will assume the first arg
// 	  is a subcommand. If you have options, you should parse those out first with ParseArgs.
// 	- default: The default subcommand to run if none is found.
// 	- offset: By default, this assumes an os.Args, and looks up the item in os.Args[1]. You can
// 	  override this behavior by setting offset to something else.
// 	- ignoreRoutes: A []string of routes that should not be executed.
func RunSubcommand(cxt cookoo.Context, params *cookoo.Params) (interface{}, cookoo.Interrupt) {
	params.Requires("args")

	args := params.Get("args", nil).([]string)
	offset := params.Get("offset", 1).(int)
	var route string
	if len(args) <= offset {
		route = params.Get("default", "default").(string)
	} else {
		route = args[offset]
	}

	stoplist := params.Get("ignoreRoutes", []string{}).([]string)
	if len(stoplist) > 0 {
		for _, stop := range stoplist {
			if stop == route {
				return nil, &cookoo.FatalError{"Illegal route."}
			}
		}
	}

	return nil, &cookoo.Reroute{route}
}

// Shift the args N (default 1) times, returning the last shifted value.
//
// Params:
// 	- n: The number of times to shift. Only the last value is returned.
// 	- args: The name of the context slice/array to modify. This value will be retrieved
// 	 from the context. Default: "os.Args"
func ShiftArgs(c cookoo.Context, params *cookoo.Params) (interface{}, cookoo.Interrupt) {

	n := params.Get("n", 1).(int)
	argName := params.Get("args", "os.Args").(string)

	args, ok := c.Get(argName, nil).([]string)
	if !ok {
		return nil, &cookoo.FatalError{"Could not get arg out of context: No such arg name."}
	}

	if len(args) < n {
		c.Put(argName, make([]string, 0))
		//log.Printf("Not enough args in %s", argName)
		return nil, &cookoo.RecoverableError{"Not enough arguments."}
	}
	targetArg := n - 1
	shifted := args[targetArg]
	c.Put(argName, args[n:])

	return shifted, nil
}
