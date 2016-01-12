package cookoo

// LogMessage prints a message to the log.
//
// Params
//
// 	- msg: The message to print
// 	- level: The log level (default: "info")
func LogMessage(cxt Context, params *Params) (interface{}, Interrupt) {
	msg := params.Get("msg", "tick")
	level := params.Get("level", "info").(string)
	cxt.Log(level, msg)
	return nil, nil
}

// AddToContext adds all of the param name/value pairs into the context.
//
// Params
//
// 	- Any params will be added into the context.
func AddToContext(cxt Context, params *Params) (interface{}, Interrupt) {
	p := params.AsMap()
	for k, v := range p {
		cxt.Put(k, v)
	}
	return true, nil
}

// ForwardTo forwards to the given route name.
//
// To prevent possible loops or problematic re-routes, use ignoreRoutes.
//
// Params
//
// 	- route: The route to forward to. This is required.
// 	- ignoreRoutes: Route names that should be ignored (generate recoverable errors).
func ForwardTo(cxt Context, params *Params) (interface{}, Interrupt) {
	ok, _ := params.Requires("route")

	if !ok {
		return nil, &FatalError{"Expected a 'route'"}
	}

	route := params.Get("route", "default").(string)

	stoplist := params.Get("ignoreRoutes", []string{}).([]string)
	if len(stoplist) > 0 {
		for _, stop := range stoplist {
			if stop == route {
				return nil, &RecoverableError{"Ignored route " + route}
			}
		}
	}

	return nil, &Reroute{route}
}
