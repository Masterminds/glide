package web

import (
	"fmt"
	"github.com/Masterminds/cookoo"
	"testing"
)

func TestUriPathResolver(t *testing.T) {
	reg, router, cxt := cookoo.Cookoo()

	resolver := NewURIPathResolver(reg)
	router.SetRequestResolver(resolver)

	// ORDER IS IMPORTANT!
	reg.Route("/foo/bar/baz", "test")
	reg.Route("/foo/bar/*", "test")
	reg.Route("/foo/c??/baz", "test")
	reg.Route("/foo/[cft]ar/baz", "test")
	reg.Route("/foo/*/baz", "test")
	reg.Route("/foo/[0-9]*/baz", "test")
	reg.Route("/*/*/*", "test")

	reg.Route("GET /foo/bar/baz", "Test with verb")
	reg.Route("POST /foo/bar/baz", "Test with verb")
	reg.Route("DELETE /foo/bar/baz", "Test with verb")
	reg.Route("* /foo/bar/baz", "Test with verb")
	reg.Route("* /foo/last", "Test with verb")

	reg.Route("GET /assets/**", "Test with double wildcard")
	reg.Route("POST /assets/foo/bar/**", "Test with double wildcard")
	reg.Route("DELETE /foo/assets/**", "Test with double wildcard")
	reg.Route("* /foo/**", "Test with double wildcard")
	reg.Route("GET /**", "Match with double wildcard")
	reg.Route("PROPFIND /assets/z/*/a/**", "Mixed wildcards")

	reg.Route("**", "Match anything")

	tests := map[string]string{
		// No Verb
		"/foo/bar/baz": "/foo/bar/baz",
		"/foo/bar/blurp": "/foo/bar/*",
		"/foo/car/baz":  "/foo/c??/baz",
		"/foo/anything/baz": "/foo/*/baz",
		"/foo/far/baz": "/foo/[cft]ar/baz",

		// Verb
		"POST /foo/bar/baz": "POST /foo/bar/baz",
		"GET /foo/last": "* /foo/last",

		// Special Wildcards
		"GET /foo/assets/img/foo.jpg": "* /foo/**",
		"DELETE /foo/assets/img/foo.png": "DELETE /foo/assets/**",
		"POST /assets/foo/bar/baz/bing/a/b/c/d/e.mov": "POST /assets/foo/bar/**",
		"GET /assets/a/b/c/d/e/f/g/h/i": "GET /assets/**",
		"GET /zzzz/a/b/c/d/e/fgh": "GET /**",
		"PROPFIND /assets/z/foo/a/b/c/d": "PROPFIND /assets/z/*/a/**",

		"STINKY /cheese/is/excellent": "**",
	}

	for name, expects := range tests {
		resolved, err := router.ResolveRequest(name, cxt)
		if err != nil {
			t.Errorf("Unexpected resolver error: %s", err)
		}
		if resolved != expects {
			t.Error(fmt.Sprintf("! Expected `%s` to match `%s`; got `%s`", name, expects, resolved))
		}
	}
}
