package main

import (
	// This is the path to Cookoo
	"fmt"
	"github.com/Masterminds/cookoo"
)

func main() {

	// Build a new Cookoo app.
	registry, router, context := cookoo.Cookoo()

	// Fill the registry.
	registry.AddRoutes(
		cookoo.Route{
			Name: "TEST",
			Help: "A test route",
			Does: cookoo.Tasks{
				cookoo.Cmd{
					Name: "hi",
					Fn:   HelloWorld,
				},
			},
		},
	)

	// Execute the route.
	router.HandleRequest("TEST", context, false)
}

func HelloWorld(cxt cookoo.Context, params *cookoo.Params) (interface{}, cookoo.Interrupt) {
	fmt.Println("Hello World")
	return true, nil
}
