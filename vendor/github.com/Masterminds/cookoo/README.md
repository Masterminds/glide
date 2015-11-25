Cookoo
======

A chain-of-command framework written in Go

[![Build Status](https://travis-ci.org/Masterminds/cookoo.png?branch=master)](https://travis-ci.org/Masterminds/cookoo) [![GoDoc](https://godoc.org/github.com/Masterminds/cookoo?status.png)](https://godoc.org/github.com/Masterminds/cookoo)

## Usage

```
$ cd $GOPATH
$ go get github.com/Masterminds/cookoo
```

Use it as follows (from `example/example.go`):

~~~go
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
~~~

## Documentation

- The [Web App tutorial](https://github.com/Masterminds/cookoo-web-tutorial)
- The [CLI Tutorial](https://github.com/Masterminds/cookoo-cli-tutorial)
- The [API
  Reference](https://godoc.org/github.com/Masterminds/cookoo)

## A Real Example

For a real example of Cookoo, take a look at
[Skunk](https://github.com/technosophos/Skunk).

Here's what Skunk's registry looks like:

```go
	registry.
	Route("scaffold", "Scaffold a new app.").
		Does(LoadSettings, "settings").
			Using("file").WithDefault(homedir + "/settings.json").From("cxt:SettingsFile").
		Does(MakeDirectories, "dirs").
			Using("basedir").From("cxt:basedir").
			Using("directories").From("cxt:directories").
		Does(RenderTemplates, "template").
			Using("tpldir").From("cxt:homedir").
			Using("basedir").From("cxt:basedir").
			Using("templates").From("cxt:templates").
	Route("help", "Print help").
		Does(Usage, "Testing")
```

This has two routes:

- scaffold
- help

The `help` route just runs the command `Usage`, which looks like this:

```go
func Usage(cxt cookoo.Context, params *cookoo.Params) interface{} {
	fmt.Println("Usage: skunk PROJECTNAME")
	return true
}
```

That is a good example of a basic command.

The `scaffold` route is more complex. It performs the following commands
(in order):

- LoadSettings: Load a `settings.json` file into the context.
- MakeDirectories: Make a bunch of directories.
- RenderTemplates: Perform template conversions on some files.

The `MakeDirectories` command is an example of a more complex command.
It takes two parameters (declared with `Using().From()`):

1. basedir: The base directory where the new subdirectories will be
   created. This comes from the `cxt:basedir` source, which means Cookoo
   looks in the `Context` object for a value named `basedir`.
2. directoies: An array of directory names that this command will
   create. These come from `cxt:directories`, which means that the
   `Context` object is queried for the value of `directories`. In this
   case, that value is actually loaded from the `settings.json` file into
   the context by the `LoadSettings` command.`

With that in mind, let's look at the command:

```go
// The MakeDirectories command.
// All commands take a Context and a Params object, and return an
// interface{}
func MakeDirectories(cxt cookoo.Context, params *cookoo.Params) interface{} {

	// This is how we get something out of the Params object. This is the
	// value that was passed in by `Using('basedir').From('cxt:basedir')
	basedir := params.Get("basedir", ".").(string)

	// This is another way to get a parameter value. This form allows us
	// to conveniently check that the parameter exists.
	d, ok := params.Has("directories")
	if !ok {
		// Did nothing. But we don't want to raise an error.
		return false
	}

	// We do have to do an explicit type conversion.
	directories := d.([]interface{})

	// Here we do the work of creating directories.
	for _, dir := range directories {
		dname := path.Join(basedir, dir.(string))
		os.MkdirAll(dname, 0755)
	}

	// We don't really have anything special to return, so we just
	// indicate that the command was successful.
	return true
}
```

This is a basic example of working with Cookoo. But far more
sophisticated workflows can be built inexpensively and quickly, and in a
style that encourages building small and re-usable chunks of code.
