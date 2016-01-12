package web

import (
	"github.com/Masterminds/cookoo"
	"net/http"
	"runtime"

	"os"
	"os/signal"
)

// Serve creates a new Cookoo web server.
//
// Important details:
//
// 	- A URIPathResolver is used for resolving request names.
// 	- The following datasources are added to the Context:
// 	  * url: A URLDatasource (Provides access to parts of the URL)
// 	  * path: A PathDatasource (Provides access to parts of a path. E.g. "/foo/bar")
// 	  * query: A QueryParameterDatasource (Provides access to URL query parameters.)
// 	  * post: A FormValuesDatasource (Provides access to form data or the body of a request.)
// 	- The following context variables are set:
// 	  * http.Request: A pointer to the http.Request object
// 	  * http.ResponseWriter: The response writer.
// 	  * server.Address: The server's address and port (NOT ALWAYS PRESENT)
// 	- The handler includes logic to redirect "not found" errors to a path named "@404" if present.
//
// Context Params:
//
// 	- server.Address: If this key exists in the context, it will be used to determine the host/port the
//   server runes on. EXPERIMENTAL. Default is ":8080".
//
// Example:
//
//    package main
//
//    import (
//      //This is the path to Cookoo
//      "github.com/Masterminds/cookoo"
//      "github.com/Masterminds/cookoo/web"
//      "fmt"
//    )
//
//    func main() {
//      // Build a new Cookoo app.
//      registry, router, context := cookoo.Cookoo()
//
//      // Fill the registry.
//      registry.Route("GET /", "The index").Does(web.Flush, "example").
//        Using("content").WithDefault("Hello World")
//
//      // Create a server
//      web.Serve(reg, router, cookoo.SyncContext(cxt))
//    }
//
// Note that we synchronize the context before passing it into Serve(). This
// is optional because each handler gets its own copy of the context already.
// However, if commands pass the context to goroutines, the context ought to be
// synchronized to avoid race conditions.
//
// Note that copies of the context are not synchronized with each other.
// So by declaring the context synchronized here, you
// are not therefore synchronizing across handlers.
func Serve(reg *cookoo.Registry, router *cookoo.Router, cxt cookoo.Context) {
	addr := cxt.Get("server.Address", ":8080").(string)

	handler := NewCookooHandler(reg, router, cxt)

	// MPB: I dont think there's any real point in having a multiplexer in
	// this particular case. The Cookoo handler is mux enough.
	//
	// Note that we can always use Cookoo with the built-in multiplexer. It
	// just doesn't make sense if Cookoo's the only handler on the app.
	//http.Handle("/", handler)

	server := &http.Server{Addr: addr}

	// Instead of mux, set a single default handler.
	// What we might be losing:
	// - Handling of non-conforming paths.
	server.Handler = handler

	go handleSignals(router, cxt, server)
	err := server.ListenAndServe()
	//err := http.ListenAndServe(addr, nil)
	if err != nil {
		cxt.Logf("error", "Caught error while serving: %s", err)
		if router.HasRoute("@crash") {
			router.HandleRequest("@crash", cxt, false)
		}
	}
}

// ServeTLS does the same as Serve, but with SSL support.
//
// If `server.Address` is not found in the context, the default address is
// `:4433`.
//
// Neither certFile nor keyFile are stored in the context. These values are
// considered to be security sensitive.
func ServeTLS(reg *cookoo.Registry, router *cookoo.Router, cxt cookoo.Context, certFile, keyFile string) {
	addr := cxt.Get("server.Address", ":4433").(string)

	server := &http.Server{Addr: addr}
	server.Handler = NewCookooHandler(reg, router, cxt)

	go handleSignals(router, cxt, server)
	err := server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		cxt.Logf("error", "Caught error while serving: %s", err)
		if router.HasRoute("@crash") {
			router.HandleRequest("@crash", cxt, false)
		}
	}
}

// handleSignals traps kill and interrupt signals and runs shutdown().
func handleSignals(router *cookoo.Router, cxt cookoo.Context, server *http.Server) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Kill, os.Interrupt)

	s := <-sig
	cxt.Logf("info", "Received signal %s. Shutting down.", s)
	// Not particularly useful on its own.
	// server.SetKeepAlivesEnabled(false)
	// TODO: Implement graceful shutdowns.
	shutdown(router, cxt)
	os.Exit(0)

}

// shutdown runs an @shutdown route if it's found in the router.
func shutdown(router *cookoo.Router, cxt cookoo.Context) {
	if router.HasRoute("@shutdown") {
		cxt.Logf("info", "Executing route @shutdown")
		router.HandleRequest("@shutdown", cxt, false)
	}
}

// The handler for Cookoo.
// You way use this handler in your own web apps, or you can use
// the Serve() function to create and manage a handler for you.
type CookooHandler struct {
	Registry    *cookoo.Registry
	Router      *cookoo.Router
	BaseContext cookoo.Context
}

// Create a new Cookoo HTTP handler.
//
// This will create an HTTP hanlder, but will not automatically attach it to a server. Implementors
// can take the handler and attach it to an existing HTTP server wiht http.HandleFunc() or
// http.ListenAndServe().
//
// For simple web servers, using this package's Serve() function may be the easier route.
//
// Important details:
//
// - A URIPathResolver is used for resolving request names.
// - The following datasources are added to the Context:
//   * url: A URLDatasource (Provides access to parts of the URL)
//   * path: A PathDatasource (Provides access to parts of a path. E.g. "/foo/bar")
//   * query: A QueryParameterDatasource (Provides access to URL query parameters.)
//   * post: A FormValuesDatasource (Provides access to form data or the body of a request.)
// - The following context variables are set:
//   * http.Request: A pointer to the http.Request object
//   * http.ResponseWriter: The response writer.
//   * server.Address: The server's address and port (NOT ALWAYS PRESENT)
func NewCookooHandler(reg *cookoo.Registry, router *cookoo.Router, cxt cookoo.Context) *CookooHandler {
	handler := new(CookooHandler)
	handler.Registry = reg
	handler.Router = router
	handler.BaseContext = cxt

	// Use the URI oriented request resolver in this package.
	resolver := new(URIPathResolver)
	resolver.Init(reg)
	router.SetRequestResolver(resolver)

	return handler
}

// Adds the built-in HTTP-specific datasources.
func (h *CookooHandler) addDatasources(cxt cookoo.Context, req *http.Request) {
	parsedURL := req.URL
	urlDS := new(URLDatasource).Init(parsedURL)
	queryDS := new(QueryParameterDatasource).Init(parsedURL.Query())
	formDS := new(FormValuesDatasource).Init(req)
	pathDS := new(PathDatasource).Init(parsedURL.Path)
	headerDS := new(RequestHeaderDatasource).Init(req)

	cxt.AddDatasource("url", urlDS)
	cxt.AddDatasource("query", queryDS)
	// cxt.AddDatasource("q", queryDS)
	cxt.AddDatasource("post", formDS)
	cxt.AddDatasource("path", pathDS)
	cxt.AddDatasource("header", headerDS)
}

// ServeHTTP is the Cookoo request handling function.
//
// This is capable of handling HTTP and HTTPS requests.
func (h *CookooHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// First we need to clone the context so we have a mutable copy.
	cxt := h.BaseContext.Copy()
	// Trap panics and make them 500 errors:
	defer func() {
		// fmt.Printf("Deferred function executed for path %s\n", req.URL.Path)
		if err := recover(); err != nil {
			//log.Printf("FOUND ERROR: %v", err)
			where := cxt.Get("command.Name", "<unknown>").(string)
			rname := cxt.Get("route.Name", "<unknown>").(string)
			h.BaseContext.Logf("error", "CookooHandler trapped a panic on route '%s' in command '%s': %v", rname, where, err)

			// Buffer for a stack trace.
			// This is pretty much always worthless, as the stack has been
			// unwound up to here.
			stack := make([]byte, 8192)
			size := runtime.Stack(stack, false)
			h.BaseContext.Logf("error", "Stack: %s", stack[:size])

			if size == 8192 {
				h.BaseContext.Logf("error", "<truncated stack trace at 8192 bytes>")
			}

			http.Error(res, "An internal error occurred.", http.StatusInternalServerError)
		}
	}()

	cxt.Put("http.Request", req)
	cxt.Put("http.ResponseWriter", res)

	// Next, we add the datasources for URL and Query params.
	h.addDatasources(cxt, req)

	// Find the route
	path := req.Method + " " + req.URL.Path

	cxt.Logf("info", "Handling request for %s\n", path)

	// If a route matches, run it.
	err := h.Router.HandleRequest(path, cxt, true)
	if err != nil {
		switch err.(type) {

		// For a 404, we bail.
		case *cookoo.RouteError:
			cxt.Logf("info", "(recovering) RouteError on route %s: %s", path, err)
			if h.Router.HasRoute("@404") {
				h.Router.HandleRequest("@404", cxt, false)
			} else {
				http.NotFound(res, req)
			}
			return
		// For any other, we go to a 500.
		case *cookoo.FatalError:
			cxt.Logf("error", "Fatal Error on route '%s': %s", path, err)
		default:
			cxt.Logf("error", "Untagged error on route '%s': %v (%T)", path, err, err)
		}

		if h.Router.HasRoute("@500") {
			cxt.Put("error", err)
			h.Router.HandleRequest("@500", cxt, false)
		} else {
			// Passing the error back to the client is a bad default.
			//http.Error(res, err.Error(), http.StatusInternalServerError)
			http.Error(res, "Internal error processing the request.", http.StatusInternalServerError)
		}
	}
}
