/* Examples and documentation for Cookoo.
*/
package main

import (
	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/cookoo/web"
	"github.com/Masterminds/cookoo/fmt"
)

// main is an example of a simple Web server written in Cookoo.
func main() {
	// First, we create a new Cookoo app.
	reg, router, cxt := cookoo.Cookoo()

	// We declare a route that answers GET requests for the path /
	//
	// By default, this will be running on http://localhost:8080/
	reg.Route("GET /", "Simple test route.").
		Does(web.Flush, "out").
			Using("content").WithDefault("OH HAI!")

	// We declare a route that answers GET requests for the path /test
	// This one uses a basic template.
	//
	// By default, this will be running on http://localhost:8080/
	//
	// Because we use `query:you`, try hitting the app on this URL:
	// http://localhost:8080/test?you=Matt
	reg.Route("GET /test", "Simple test route.").
		Does(fmt.Template, "content").
			Using("template").WithDefault("Hello {{.you}}").
		Using("you").WithDefault("test").From("query:you").
		Does(web.Flush, "out").
			Using("content").From("cxt:content")

	// Start the server.
	web.Serve(reg, router, cxt)
}
