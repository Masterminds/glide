package cookoo

import (
	"fmt"
	"testing"
)

func Example() {
	// This is an admittedly contrived example in which we first store a
	// "Hello World" message, and then tell the logger to get that stored
	// message and write it to the log.
	reg, router, cxt := Cookoo()
	reg.AddRoute(Route{
		Name: "hello",
		Help: "Sends the log message 'Hello World'",
		Does: Tasks{
			// First, store the message "Hello World" in the context.
			Cmd{
				Name: "message",
				Fn:   AddToContext,
				Using: []Param{
					Param{
						Name:         "hello",
						DefaultValue: "Hello World",
					},
				},
			},
			// Now get that message and write it to the log.
			Cmd{
				Name: "log",
				Fn:   LogMessage,
				Using: []Param{
					Param{
						Name: "msg",
						From: "cxt:message",
					},
				},
			},
		},
	})

	router.HandleRequest("hello", cxt, false)
}

func ExampleCookoo() {
	reg, router, cxt := Cookoo()
	reg.AddRoute(Route{
		// The name of the route. You execute routes by name. (See router.HandleRequest below)
		Name: "hello",
		// This is for documentation/help tools.
		Help: "Print a message on standard output",

		// This is a list of things you want this route to do. When executed,
		// it will run these commands in order.
		Does: Tasks{
			// Declare a new command.
			Cmd{
				// Give the command a name. Programs reference command output
				// by this name.
				Name: "print",

				// Tell Cookoo what function to execute when we get to this
				// step.
				//
				// Usually we define functions elsewhere so we can re-use them.
				Fn: func(c Context, p *Params) (interface{}, Interrupt) {
					// Print whatever the content of the 'msg' parameter is.
					fmt.Println(p.Get("msg", "").(string))
					return nil, nil
				},
				// Send some parameters into Fn. Here we define the 'msg'
				// parameter that Fn prints. While we just use a default
				// value here, Cookoo can get that information from another
				// source and then send it into Fn.
				Using: []Param{
					Param{
						Name:         "msg",
						DefaultValue: "Hello World",
					},
				},
			},
		},
	})

	// Now we execute the "hello" chain of commands.
	router.HandleRequest("hello", cxt, false)
	// Output:
	// Hello World
}

func TestCookooForCoCo(t *testing.T) {
	registry, router, cxt := Cookoo()

	cxt.Put("Answer", 42)

	lifeUniverseEverything := cxt.Get("Answer", nil)

	if lifeUniverseEverything != 42 {
		t.Error("! Context is not working.")
	}

	registry.Route("foo", "test")

	ok := router.HasRoute("foo")

	if !ok {
		t.Error("! Router does not have 'foo' route.")
	}
}
