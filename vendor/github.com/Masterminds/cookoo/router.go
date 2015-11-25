package cookoo

import (
	"fmt"
	"strings"
)

// RequestResolver is the interface for the request resolver.
// A request resolver is responsible for transforming a request name to
// a route name. For example, a web-specific resolver may take a URI
// and return a route name. Or it make take an HTTP verb and return a
// route name.
type RequestResolver interface {
	Init(registry *Registry)
	Resolve(path string, cxt Context) (string, error)
}

// Router is the Cookoo router.
// A Cookoo app works by passing a request into a router, and
// relying on the router to execute the appropriate chain of
// commands.
type Router struct {
	registry *Registry
	resolver RequestResolver
}

// BasicRequestResolver is a basic resolver that assumes that the given request
// name *is* the route name.
type BasicRequestResolver struct {
	registry *Registry
	resolver RequestResolver
}

// NewRouter creates a new Router.
func NewRouter(reg *Registry) *Router {
	router := new(Router)
	router.Init(reg)
	return router
}

// Init initializes the BasicRequestResolver.
func (r *BasicRequestResolver) Init(registry *Registry) {
	r.registry = registry
}

// Resolve returns the given path.
// This is a non-transforming resolver.
func (r *BasicRequestResolver) Resolve(path string, cxt Context) (string, error) {
	return path, nil
}

// Init initializes the Router.
func (r *Router) Init(registry *Registry) *Router {
	r.registry = registry
	r.resolver = new(BasicRequestResolver)
	r.resolver.Init(registry)
	return r
}

// SetRegistry sets the registry.
func (r *Router) SetRegistry(reg *Registry) {
	r.registry = reg
}

// SetRequestResolver sets the request resolver.
// The resolver is responsible for taking an arbitrary string and
// resolving it to a registry route.
//
// Example: Take a URI and translate it to a route.
func (r *Router) SetRequestResolver(resolver RequestResolver) {
	r.resolver = resolver
}

// RequestResolver gets the request resolver.
func (r *Router) RequestResolver() RequestResolver {
	return r.resolver
}

// ResolveRequest resolver a given string into a route name.
func (r *Router) ResolveRequest(name string, cxt Context) (string, error) {
	routeName, e := r.resolver.Resolve(name, cxt)

	if e != nil {
		return routeName, e
	}

	return routeName, nil
}

// HandleRequest does a request.
// This executes a request "named" name (this string is passed through the
// request resolver.) The context is cloned (shallow copy) and passed in as the
// base context.
//
// If taint is `true`, then no routes that begin with `@` can be executed. Taint
// should be set to true on anything that relies on a name supplied by an
// external client.
//
// This will do the following:
// 	- resolve the request name into a route name (using a RequestResolver)
// 	- look up the route
// 	- execute each command on the route in order
//
// The following context variables are placed into the context during a run:
//
// 	route.Name - Processed name of the current route
// 	route.Description - Description of the current route
// 	route.RequestName - raw route name as passed by the client
// 	command.Name - current command name (changed with each command)
//
// If an error occurred during processing, an error type is returned.
func (r *Router) HandleRequest(name string, cxt Context, taint bool) error {

	// Not sure why we were passing a copy of the context?
	// baseCxt := cxt.Copy()
	// routeName, e := r.ResolveRequest(name, baseCxt)
	routeName, e := r.ResolveRequest(name, cxt)

	if e != nil {
		return e
	}

	cxt.Put("route.RequestName", name)
	cxt.Put("route.Name", routeName)
	if spec, ok := r.registry.RouteSpec(routeName); ok {
		cxt.Put("route.Description", spec.description)
	}

	// Let an outer routine call go HandleRequest()
	//go r.runRoute(routeName, cxt, taint)
	e = r.runRoute(routeName, cxt, taint)

	return e
}

// HasRoute checks whether or not the route exists.
// Note that this does NOT resolve a request name into a route name. This
// expects a route name.
func (r *Router) HasRoute(name string) bool {
	_, ok := r.registry.RouteSpec(name)
	return ok
}

// PRIVATE ==========================================================

// Given a router, context, and taint, run the route.
func (r *Router) runRoute(route string, cxt Context, taint bool) error {
	if len(route) == 0 {
		return &RouteError{"Empty route name."}
	}
	if taint && route[0] == '@' {
		return &RouteError{"Route is tainted. Refusing to run."}
	}
	spec, ok := r.registry.RouteSpec(route)
	if !ok {
		return &RouteError{fmt.Sprintf("Route %s does not exist.", route)}
	}
	// fmt.Printf("Running route %s: %s\n", spec.name, spec.description)
	for _, cmd := range spec.commands {
		// Provide info for each run.
		cxt.Put("command.Name", cmd.name)

		// fmt.Printf("Command %d is %s (%T)\n", i, cmd.name, cmd.command)
		res, irq := r.doCommand(cmd, cxt)

		// This may store a nil.
		cxt.Put(cmd.name, res)

		// Handle interrupts.
		if irq != nil {
			// If this is a reroute, call runRoute() again.
			reroute, isType := irq.(*Reroute)
			if isType {
				routeName, e := r.ResolveRequest(reroute.RouteTo(), cxt)
				if e != nil {
					return e
				}
				//fmt.Printf("Routing to %s\n", routeName)
				// MPB: I think re-routes should disable taint mode, since they
				// are explicitly called from within the code.
				return r.runRoute(routeName, cxt /*taint*/, false)
			}

			_, isType = irq.(*Stop)
			if isType {
				return nil
			}

			// If this is a recoverable error, recover and go on.
			err, isType := irq.(*RecoverableError)
			// Otherwise, terminate the route.
			if isType {
				// Swallow the error.
				// XXX: Should this be logged?
				cxt.Logf("warn", "Continuing after Recoverable Error on route %s: %v", route, err)
			} else {
				// return irq.(*FatalError)
				return irq.(error)
			}
		}
	}
	return nil
}

// Do an individual command.
func (r *Router) doCommand(cmd *commandSpec, cxt Context) (interface{}, Interrupt) {
	params := r.resolveParams(cmd, cxt)

	ret, irq := cmd.command(cxt, params)
	return ret, irq
}

// Get the appropriate values for each param.
func (r *Router) resolveParams(cmd *commandSpec, cxt Context) *Params {
	parameters := NewParams(len(cmd.parameters))
	for _, ps := range cmd.parameters {
		sources := parseFromStatement(ps.from)
		val := r.defaultFromSources(sources, cxt)
		if val == nil {
			parameters.set(ps.name, ps.defaultValue)
			val = ps.defaultValue
		}
		parameters.set(ps.name, val)
	}
	return parameters
}

// Get the values from a source.
// Returns the value of the first source to return a non-nil value.
func (r *Router) defaultFromSources(sources []*fromVal, cxt Context) interface{} {
	for _, src := range sources {
		switch src.source {
		case "c", "cxt", "context":
			val, ok := cxt.Has(src.key)
			if ok {
				return val
			}
		case "datasource", "ds":
			ds, ok := cxt.HasDatasource(src.key)
			if ok {
				return ds
			}
		default:
			// If we have a datasource, and the datasource
			// is a KeyValueDatasource, try to return the value.
			if ds, ok := cxt.HasDatasource(src.source); ok {
				store, ok := ds.(KeyValueDatasource)
				if ok {
					v := store.Value(src.key)
					if v != nil {
						return v
					}
					//fmt.Printf("V is nil for %v\n", src)
				}
			}
		}
	}
	return nil
}

// Parse a 'from' statement.
func parseFromStatement(from string) []*fromVal {
	toks := strings.Fields(from)
	ret := make([]*fromVal, len(toks))
	for i, tok := range toks {
		ret[i] = parseFromVal(tok)
	}
	return ret
}

// Represents a 'from' value of a 'from' statement.
type fromVal struct {
	source, key string
}

// Parse a FROM string of the form NAME:VALUE
func parseFromVal(from string) *fromVal {
	vals := strings.SplitN(strings.TrimSpace(from), ":", 2)
	if len(vals) == 1 {
		return &fromVal{vals[0], ""}
	}
	return &fromVal{vals[0], vals[1]}
}

// RouteError indicates that a route cannot be executed successfully.
type RouteError struct {
	Message string
}

// Error returns the error.
func (e *RouteError) Error() string {
	return e.Message
}
