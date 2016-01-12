# Changelog

## v1.2.0 (2015-06-16)

* EXPERIMENTAL: Added support for CmdDef, a tool for writing structs
  that manage types for cookoo commands.
* Added support for an alternative Route declaration syntax that does
  not use method chaining.
* Added support for @shutdown routes on web.Serve().
* Added safely.GoDo(cxt, GoDoer)
* NewReroute(route string) can be used to create Reroutes now.
* From() now takes an vararg of strings: `From("cxt:foo", "cxt:bar")`.
  It can still take a space-delimited set of strings, too. In fact, both
  can be used together:`From("cxt:foo", "cxt:bar cxt:baz")`.
* Added Getter interface
* Added the Get* and Has* utility functions (getter.go)
* Added GetFromFirst(string, interface) (Contextvalue, Getter) function
* Added DefaultGetter struct
* Added 'subcommand' param to cli.ParseArgs
* Added 'cli.New' and 'cli.Runner'
* Added 'fmt' package

## v1.1.0 (2014-06-06)

* Added SyncContext function so Contexts are kept in sync via read/write mutex.
* Improved debugging for panics.
* Documentation fixes.

## v1.0.0 (2014-04-23)

* Initial release.
