package cookoo

// Copyright 2013 Masterminds.

import (
	"fmt"
	"strings"
)

// A Registry contains the the callback routes and the commands each
// route executes.
type Registry struct {
	routes            map[string]*routeSpec
	orderedRouteNames []string
	currentRoute      *routeSpec
}

// NewRegistry returns a new initialized registry.
func NewRegistry() *Registry {
	r := new(Registry)
	r.Init()
	return r
}

// Init initializes a registry. If a Registry is created through a means other
// than NewRegistry Init should be called on it.
func (r *Registry) Init() *Registry {
	// Why 8?
	r.routes = make(map[string]*routeSpec, 8)
	r.orderedRouteNames = make([]string, 0, 8)
	return r
}

// Route specifies a new route to add to the registry.
func (r *Registry) Route(name, description string) *Registry {

	// Create the route spec.
	route := new(routeSpec)
	route.name = name
	route.description = description
	route.commands = make([]*commandSpec, 0, 4)

	// Add the route spec.
	r.currentRoute = route
	r.routes[name] = route
	r.orderedRouteNames = append(r.orderedRouteNames, name)

	return r
}

func (r *Registry) DoesCmdDef(cd CommandDefinition, name string) *Registry {
	// Configure command spec.
	spec := new(commandSpec)
	spec.name = name
	spec.command = func(c Context, p *Params) (interface{}, Interrupt) {
		// We don't have to clone cmd.Def because Map builds
		// a new copy.
		o, err := Map(c, p, cd)
		if err != nil {
			return nil, err
		}
		return o.Run(c)
	}

	// Add command spec.
	r.currentRoute.commands = append(r.currentRoute.commands, spec)

	return r
}

// Does adds a command to the end of the chain of commands for the current
// (most recently specified) route.
func (r *Registry) Does(cmd Command, commandName string) *Registry {

	// Configure command spec.
	spec := new(commandSpec)
	spec.name = commandName
	spec.command = cmd

	// Add command spec.
	r.currentRoute.commands = append(r.currentRoute.commands, spec)

	return r
}

// Using specifies a paramater to use for the most recently specified command
// as set by Does.
func (r *Registry) Using(name string) *Registry {
	// Look up the last command added.
	lastCommand := r.lastCommandAdded()

	// Create a new spec.
	spec := new(paramSpec)
	spec.name = name

	// Add it to the list.
	lastCommand.parameters = append(lastCommand.parameters, spec)
	return r
}

// WithDefault specifies the default value for the most recently specified
// parameter as set by Using.
func (r *Registry) WithDefault(value interface{}) *Registry {
	param := r.lastParamAdded()
	param.defaultValue = value
	return r
}

// From sepcifies where to get the value from for the most recently specified
// paramater as set by Using.
func (r *Registry) From(fromVal ...string) *Registry {
	param := r.lastParamAdded()

	// This is sort of a hack. Really, we should make params.from a []string.
	param.from = strings.Join(fromVal, " ")
	return r
}

// Get the last parameter for the last command added.
func (r *Registry) lastParamAdded() *paramSpec {
	cspec := r.lastCommandAdded()
	last := len(cspec.parameters) - 1
	return cspec.parameters[last]
}

// Includes makes the commands from another route avaiable on this route.
func (r *Registry) Includes(route string) *Registry {

	// Not that we don't clone commands; we just add the pointer to the current
	// route.
	spec := r.routes[route]
	if spec == nil {
		panicString := fmt.Sprintf("Could not find route %s. Skipping include.", route)
		panic(panicString)
	}
	for _, cmd := range spec.commands {
		r.currentRoute.commands = append(r.currentRoute.commands, cmd)
	}
	return r
}

// RouteSpec gets a ruote cased on its name.
func (r *Registry) RouteSpec(routeName string) (spec *routeSpec, ok bool) {
	spec, ok = r.routes[routeName]
	return
}

// Routes gets an unordered map of routes names to route specs.
//
// If order is important, use RouteNames to get the names (in order).
func (r *Registry) Routes() map[string]*routeSpec {
	return r.routes
}

// RouteNames gets a slice containing the names of every registered route.
//
// The route names are returned in the order they were added to the
// registry. This is useful to some resolvers, which apply rules in order.
func (r *Registry) RouteNames() []string {
	return r.orderedRouteNames
	/*
		names := make([]string, len(r.routes))
		i := 0
		for k := range r.routes {
			names[i] = k
			i++
		}
		return names
	*/
}

// Look up the last command.
func (r *Registry) lastCommandAdded() *commandSpec {
	lastIndex := len(r.currentRoute.commands) - 1
	return r.currentRoute.commands[lastIndex]
}

type RouteDetails interface {
	Name() string
	Description() string
}

type routeSpec struct {
	name, description string
	commands          []*commandSpec
}

func (r *routeSpec) Name() string {
	return r.name
}

func (r *routeSpec) Description() string {
	return r.description
}

type commandSpec struct {
	name       string
	command    Command
	parameters []*paramSpec
}

type paramSpec struct {
	name         string
	defaultValue interface{}
	from         string
}

// New public API

// AddRoute adds a single route to the registry.
func (r *Registry) AddRoute(route Route) error {
	return r.AddRoutes(route)
}

func extractParams(cmd Task) []*paramSpec {
	var paramspecs []*paramSpec

	// FIXME: This is a horrible way to do this.
	paramspecs = make([]*paramSpec, len(cmd.getParams()))
	for j, prm := range cmd.getParams() {
		pspec := &paramSpec{
			name:         prm.Name,
			defaultValue: prm.DefaultValue,
			from:         prm.From,
		}
		paramspecs[j] = pspec
	}
	return paramspecs
}

// AddRoutes adds one or more routes to the registry.
func (r *Registry) AddRoutes(routes ...Route) error {
	for _, route := range routes {

		cmdspecs := make([]*commandSpec, 0, len(route.Does))
		for _, cmd := range route.Does {
			switch cmd := cmd.(type) {
			case CmdDef:
				// This wraps the CmdDef inside of a command.
				paramspecs := extractParams(cmd)
				cmdspec := &commandSpec{
					name: cmd.Name,
					command: func(c Context, p *Params) (interface{}, Interrupt) {
						// We don't have to clone cmd.Def because Map builds
						// a new copy.
						o, err := Map(c, p, cmd.Def)
						if err != nil {
							return nil, err
						}
						return o.Run(c)
					},
					parameters: paramspecs,
				}
				cmdspecs = append(cmdspecs, cmdspec)

			case Cmd:
				paramspecs := extractParams(cmd)

				cmdspec := &commandSpec{
					name:       cmd.Name,
					command:    cmd.Fn,
					parameters: paramspecs,
				}
				cmdspecs = append(cmdspecs, cmdspec)
			case Include:
				other, ok := r.RouteSpec(cmd.Path)
				if !ok {
					// Route not found.
					return fmt.Errorf("Route '%s' not found.", cmd.Path)
				}
				cmdspecs = append(cmdspecs, other.commands...)

			}
		}

		rspec := &routeSpec{
			name:        route.Name,
			description: route.Help,
			commands:    cmdspecs,
		}
		// Add the route spec.
		r.currentRoute = rspec
		r.routes[rspec.name] = rspec
		r.orderedRouteNames = append(r.orderedRouteNames, rspec.name)
	}
	return nil
}

// Route declares a new Cookoo route.
//
// A Route has a name, which is used to identify and call it, and Help. The
// Help can be used by other tools to generate help text or information about
// an application's structure.
//
// Routes are composed of a series of Tasks, each of which is executed in
// order.
type Route struct {
	Name, Help string
	Does       Tasks
}

// Tasks represents a list of discrete tasks that are run on a Route.
//
// There are two kinds of Tasks: Cmd (a command) and Include, which imports a
// Tasks list from another route.
type Tasks []Task

// Cmd associates a cookoo.Command to a Route.
//
// The Name is the direct reference to a command. When a Command returns output,
// that output is inserted into the Context with the key Name.
//
// Fn specifies which cookoo.Command should be executed during this step.
//
// Using contains a list of Parameters that Cookoo can pass into the Command
// at execution time.
type Cmd struct {
	Name  string
	Fn    Command
	Using Parameters
}

// Include imports all of the Tasks on another route into the present Route.
type Include struct {
	Path string
}

type CmdDef struct {
	Name  string
	Def   CommandDefinition
	Using Parameters
}

// A Task can be either an Include or a Cmd. This is a very lame way of
// making this behavior private.

type Task interface {
	getParams() Parameters
}

func (i Include) getParams() Parameters {
	return Parameters{}
}
func (c Cmd) getParams() Parameters {
	return c.Using
}
func (c CmdDef) getParams() Parameters {
	return c.Using
}

type Parameters []Param

// Param describes an individual parameter which will be passed to a Command.
//
// The Name is the name of the parameter. The Command itself dictates which
// Names it uses.
//
// The DefaultValue is the value of the Parameter if nothing else is specified.
//
// From indicates where the Param value may come from. Examples: `From("cxt:foo")`
// gets the value from the value of the key 'foo' in the Context.
type Param struct {
	Name         string
	DefaultValue interface{}
	From         string
}
