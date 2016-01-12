// Extra datasources for Web servers.
package web

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Get the query parameters by name.
type QueryParameterDatasource struct {
	Parameters url.Values
}

// The datasource for URLs.
// This datasource knows the following items:
// - url: the URL struct
// - scheme: The scheme of the URL as a string
// - opaque: The opaque identifier
// - user: A *Userinfo
// - host: The string hostname
// - path: The entire path
// - rawquery: The query string, not decoded.
// - fragment: The fragment string.
// - query: The array of Query parameters. Usually it is better to use the
//   'query:foo' syntax.
type URLDatasource struct {
	URL *url.URL
}

type RequestHeaderDatasource struct {
	req *http.Request
}

func (r *RequestHeaderDatasource) Init(req *http.Request) *RequestHeaderDatasource {
	r.req = req
	return r
}
func (r *RequestHeaderDatasource) Value(name string) interface{} {
	// We return a nil so that the context can substitute in a
	// default value. Empty string for headers seems to always mean
	// "No header found". So a default value 'bar' is what is expected when
	// From("header:foo").WithDefault("bar") and foo is not present.
	v := r.req.Header.Get(name)
	if len(v) == 0 {
		return nil
	}
	return v
}

// Access to name/value pairs in POST/PUT form data from the body.
// This will attempt to access form data supplied in the HTTP request's body.
// If the MIME type is not correct or if there is no POST data, no data will
// be made available.
//
// Parsing is lazy: No form data is parsed until it is requested.
type FormValuesDatasource struct {
	req *http.Request
}

func (f *FormValuesDatasource) Init(req *http.Request) *FormValuesDatasource {
	f.req = req
	return f
}

// The return value will always be a string or nil.
// To match the interface, we use interface{}.
func (f *FormValuesDatasource) Value(name string) interface{} {
	return f.req.PostFormValue(name)
}

func (d *QueryParameterDatasource) Init(vals url.Values) *QueryParameterDatasource {
	d.Parameters = vals
	return d
}

func (d *QueryParameterDatasource) Value(name string) interface{} {
	v := d.Parameters.Get(name)

	// We need to do this to meet the expectations of the datasource system.
	if len(v) == 0 {
		return nil
	}
	return v
}

func (d *URLDatasource) Init(parsedUrl *url.URL) *URLDatasource {
	d.URL = parsedUrl
	return d
}

func (d *URLDatasource) Value(name string) interface{} {
	switch name {
	case "host", "Host":
		return d.URL.Host
	case "path", "Path":
		return d.URL.Path
	case "url", "URL", "Url":
		return d.URL
	case "user", "User":
		return d.URL.User
	case "scheme", "Scheme":
		return d.URL.Scheme
	case "rawquery", "RawQuery":
		return d.URL.RawQuery
	case "query", "Query":
		return d.URL.Query()
	case "fragment", "Fragment":
		return d.URL.Fragment
	case "opaque", "Opaque":
		return d.URL.Opaque
	}
	return nil
}

type PathDatasource struct {
	PathParts []string
}

func (d *PathDatasource) Init(path string) *PathDatasource {
	d.PathParts = strings.Split(path, "/")[1:]
	return d
}

func (d *PathDatasource) Value(name string) interface{} {
	index, err := strconv.Atoi(name)
	if err != nil || index > len(d.PathParts) {
		return nil
	}
	return d.PathParts[index]
}

// This provides a datasource for session data.
//
// Sessions differ a little from the other web datasources in that they may
// need explicit app-controlled initialization.
type SessionDatasource interface {
	StartSession(res http.ResponseWriter, req *http.Request) bool
	ClearSession(res http.ResponseWriter, req *http.Request) bool
}

/*
type HTMLTemplateCache struct {
	//tpls map[string]*templates.Template
	tpls templates.Template
}

// Create a new template cache with the associated templates.
//
// The array of strings may be either file names or shell glob patterns. In any
// case, the pattern must match at least one file, or a fatal error will occur.
func NewHTMLTemplateCache(filenames []string) *HTMLTemplateCache {
	t := new(HTMLTemplateCache)
	tpls = templates.New("base")
	for _, filename := range filenames {
		_, err := t.ParseGlob(filename)
		if err != nil {
			panic("Could not process templates! ", err)
		}
	}
	return t
}

// Add one or more templates to the cache.
//
// filenames may be glob patterns, but at least one file must match.
//
// If a template fails to parse, an error will be returned, but this will not
// cause a panic. Any templates added here will be accessible for anything else
// that accesses the cache.
func (c *HTMLTemplateCache) AddTemplates(filenames ...string) error {
	for _, f := range filenames {
		_, e := c.tpls.ParseGlob(f)
		if e != nil {
			return e
		}
	}
	return nil
}

// Clone the existing template cache.
//
// This is useful if a particular command needs to modify the template cache
// for it's execution without changing the generally available cache.
//
// If the cache cannot be cloned, an error will be returned. This will not
// cause a panic.
func (c *HTMLTemplateCache) Clone() (*templates.Template, error) {
	return c.tpls.Clone()
}
*/
