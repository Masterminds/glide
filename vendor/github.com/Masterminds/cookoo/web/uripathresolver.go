package web

import (
	"github.com/Masterminds/cookoo"
	"path"
	"strings"
)

// Resolver for transforming a URI path into a route.
//
// This is a more sophisticated path resolver, aware of
// heirarchyand wildcards.
//
// Examples:
// - URI path `/foo` matches the entry `/foo`
// - URI path `/foo/bar` could match entries like `/foo/*`, `/foo/**`, and `/foo/bar`
// - URI path `/foo/bar/baz` could match `/foo/*/baz` and `/foo/**`
//
// HTTP Verbs:
// This resolver also allows you to specify verbs at the beginning of a path:
// - "GET /foo" and "POST /foo" are separate (but legal) paths. "* /foo" will allow any verb.
// - There are no constrainst on verb name. Thus, verbs like WebDAV's PROPSET are fine, too. Or you can
//   make up your own.
//
// IMPORTANT! When it comes to matching route patterns against paths, ORDER IS
// IMPORTANT. Routes are evaluated in order. So if two rules (/a/b* and /a/bc) are
// both defined, the incomming request /a/bc will match whichever route is
// defined first. See the unit tests for examples.
//
// The `**` and `/**` Wildcards:
// =============================
//
// In addition to the paths described in the `path` package of Go's core, two
// extra wildcard sequences are defined:
//
// - `**`: Match everything.
// - `/**`: a suffix that matches any sub-path.
//
// The `**` wildcard works in ONLY ONE WAY:  If the path is declared as `**`, with nothing else, 
// then any path will match.
//
// VALID: `**`, `GET /foo/**`, `GET /**`
// NOT VALID: `GET **`, `**/foo`, `foo/**/bar`
//
// The `/**` suffix can only be added to the end of a path, and says "Match
// anything under this".
//
// Examples:
// - URI paths "/foo", "GET /a/b/c", and "hello" all match "**". (The ** rule
//   can be very dangerous for this reason.)
// - URI path "/assets/images/foo/bar/baz.jpg" matches "/assets/**"
// 
// The behavior for rules that contain `/**` anywhere other than the end
// have undefined behavior.
//
type URIPathResolver struct {
	registry *cookoo.Registry
}

// Creates a new URIPathResolver.
func NewURIPathResolver(reg *cookoo.Registry) *URIPathResolver {
	res := new(URIPathResolver)
	res.Init(reg)
	return res
}

func (r *URIPathResolver) Init(registry *cookoo.Registry) {
	r.registry = registry
}

// Resolve a path name based using path patterns.
//
// This resolver is designed to match path-like strings to path patterns. For example,
// the path `/foo/bar/baz` may match routes like `/foo/*/baz` or `/foo/bar/*`
func (r *URIPathResolver) Resolve(pathName string, cxt cookoo.Context) (string, error) {
	// HTTP verb support naturally falls out of the fact that spaces in paths are legal in UNIXy systems, while
	// illegal in URI paths. So presently we do no special handling for verbs. Yay for simplicity.
	for _, pattern := range r.registry.RouteNames() {

		if strings.HasSuffix(pattern, "**") {
			ok := r.subtreeMatch(cxt, pathName, pattern)
			if ok {
				return pattern, nil
			}
		}

		if ok, err := path.Match(pattern, pathName); ok && err == nil {
			return pattern, nil
		} else if err != nil {
			// Bad pattern
			return pathName, err
		}
	}
	return pathName, &cookoo.RouteError{"Could not resolve route " + pathName}
}

func (r *URIPathResolver) subtreeMatch(c cookoo.Context, pathName, pattern string) bool {

	if pattern == "**" {
		return true
	}

	// Find out how many slashes we have.
	countSlash := strings.Count(pattern, "/")

	// '**' matches anything.
	if countSlash == 0 {
		c.Logf("warn", "Illegal pattern: %s", pattern)
		return false
	}

	// Add 2 for verb plus trailer.
	parts := strings.SplitN(pathName, "/", countSlash + 1)
	prefix := strings.Join(parts[0:countSlash], "/")

	subpattern := strings.Replace(pattern, "/**", "", -1)
	if ok, err := path.Match(subpattern, prefix); ok && err == nil {
		return true
	} else if err != nil {
		c.Logf("warn", "Parsing path `%s` gave error: %s", err)
	}
	return false
}
