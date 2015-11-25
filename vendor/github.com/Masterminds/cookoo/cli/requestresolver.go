package cli

import (
	"flag"
	"github.com/Masterminds/cookoo"
	//"os"
	"strings"
)


// A CLI-centered request resolver with support for command line flags.
//
// When this resolver finds "globalFlags" in the base context, it
// attempts to parse out the flags from the `path` string.
//
// You must specify arguments as a `flag.FlagSet` (see the Go documentation).
//
// Parsed commanline arguments are put directly into the context (though later we may put them
// in a datasource instead). Currently, all parameter values -- even booleans -- are stored
// as strings. If it is important to you to store them as more complex types, you may need to
// use `FlagSet.*Var()` functions.
//
// Splitting the path string into arguments is done naively by splitting the string on the
// space (%20) character.
//
// Arguments are parsed when the request resolver is asked to resolve (no earlier). Thus the FlagSet
// may be placed into the context at any time prior to cookoo.Router.HandleRequest().
//
// CONTEXT VALUES
// - globalFlags: A `flag.FlagSet`. If this is present and not empty, the path will be parsed.
type RequestResolver struct {
	registry *cookoo.Registry
}

func (r *RequestResolver) Init(registry *cookoo.Registry) {
	r.registry = registry
}

func (r *RequestResolver) Resolve(path string, cxt cookoo.Context) (string, error) {
	// Parse out any flags. Maybe flag specs are in context?

	flagsetO, ok := cxt.Has("globalFlags")
	if !ok {
		// No args to parse. Just return path.
		return path, nil
	}
	flagset := flagsetO.(*flag.FlagSet)
	flagset.Parse(strings.Split(path, " "))
	addFlagsToContext(flagset, cxt)
	args := flagset.Args()

	// This is a failure condition... Need to fix Cookoo to support error return.
	if len(args) == 0 {
		return path, &cookoo.RouteError{"Could not resolve route " + path}
	}

	// Put the rest of the args to the context.
	cxt.Put("args", args[1:])

	// Parse argv[0] as subcommand
	return args[0], nil
}

func addFlagsToContext(flagset *flag.FlagSet, cxt cookoo.Context) {
	store := func(f *flag.Flag) {
		// fmt.Printf("Storing %s in context with value %s.\n", f.Name, f.Value.String())

		// Basically, we can tell the difference between booleans and strings, and that's it.
		// Other types are a loss.
		/*
			if f.IsBoolFlag != nil {
				cxt.Put(f.Name, f.Value.String() == "true")
			} else {
				cxt.Put(f.Name, f.Value.String())
			}
		*/
		cxt.Put(f.Name, f.Value.String())
	}

	flagset.VisitAll(store)
}
