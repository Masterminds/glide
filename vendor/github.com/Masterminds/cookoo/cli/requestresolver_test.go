package cli

import (
	"flag"
	"fmt"
	"github.com/Masterminds/cookoo"
	"testing"
)

func Nothing(cxt cookoo.Context, params *cookoo.Params) (res interface{}, i cookoo.Interrupt) {
	//fmt.Printf("Nothing command was run.\n")
	return true, nil
}

func TestResolvingSimpleRoute(t *testing.T) {
	registry, router, context := cookoo.Cookoo()

	resolver := new(RequestResolver)
	resolver.Init(registry)

	router.SetRequestResolver(resolver)

	registry.Route("test", "A simple test").Does(Nothing, "nada")

	e := router.HandleRequest("test", context, false)

	if e != nil {
		t.Error("! Failed 'test' route.")
	}
}

func TestResolvingWithFlags(t *testing.T) {
	registry, router, context := cookoo.Cookoo()

	resolver := new(RequestResolver)
	resolver.Init(registry)

	flagset := flag.NewFlagSet("test", flag.ExitOnError)
	flagset.String("foo", "bar", "this is a test")
	context.Add("globalFlags", flagset)

	router.SetRequestResolver(resolver)

	registry.Route("test", "Test flag parsing.").Does(Nothing, "nada")

	e := router.HandleRequest("-foo arg1 test -foo arg2 arg3 arg4", context, false)
	if e != nil {
		t.Error("! Failed 'test' route.", e)
		return
	}

	nada, ok := context.Has("nada")
	if !ok {
		t.Error("! Expected to find a context entry for 'nada'")
		return
	}
	if !nada.(bool) {
		t.Error("! Expected 'nada' to be set to TRUE")
		return
	}

	fooArgO, ok := context.Has("foo")
	if !ok {
		t.Error("! Expected to find 'foo' in context, but it's not there.", ok)
		return
	}
	fooArg := fooArgO.(string)
	if fooArg != "arg1" {
		t.Error("! Expected 'arg1' in context 'foo'; got ", fooArg)
		return
	}

	argsListO, ok := context.Has("args")
	if !ok {
		t.Error("! Expected to find a list of args in the context as 'args'")
		return
	}

	argsList := argsListO.([]string)
	if len(argsList) != 4 {
		t.Error("! Expected to find 4 arguments in the list. Found ", len(argsList))
		t.Error(fmt.Sprintf("Arguments left: %v\n", argsList))
		return
	}

	if argsList[0] != "-foo" {
		t.Error("! Expected argList[0] to be '-foo'. Got ", argsList[0])
	}

	if argsList[3] != "arg4" {
		t.Error("! Expected argList[3] to be 'arg4'. Got ", argsList[3])
	}

}
