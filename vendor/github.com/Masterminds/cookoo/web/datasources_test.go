package web

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestURLDatasource(t *testing.T) {
	rawurl := "http://user:password@example.com/path?foo=bar#fragment"
	testUrl, err := url.Parse(rawurl)

	if err != nil {
		t.Error("! Unexpected error.", err)
		return
	}
	ds := new(URLDatasource)
	ds.Init(testUrl)

	// Test the string values.
	arr := map[string]string{
		"scheme":   "http",
		"Path":     "/path",
		"host":     "example.com",
		"fragment": "fragment",
	}
	for key, val := range arr {
		if ds.Value(key) != val {
			t.Error(fmt.Sprintf("! Expected '%s', got '%s'", val, ds.Value(key)))
		}
	}

	// Test the Query Values object.
	qvals := ds.Value("Query").(url.Values)
	if qvals.Get("foo") != "bar" {
		t.Error("! Expected to find foo=bar query param. Found ", qvals["foo"])
	}

	// Test the Userinfo object
	uinfo := ds.Value("User").(*url.Userinfo)

	if uinfo.Username() != "user" {
		t.Error("! Expected user name 'user', got ", uinfo.Username)
	}
}

func TestQueryParameterDatasource(t *testing.T) {
	testUrl, err := url.ParseRequestURI("/foo?a=b&c=foo+bar&d=1234&d=5678")
	if err != nil {
		t.Error("! Unexpected URL parse error.")
	}
	ds := new(QueryParameterDatasource).Init(testUrl.Query())

	// Test the string values.
	arr := map[string]string{
		"a": "b",
		"c": "foo bar",
		// url.Values.Get accesses the first value associated with a key. To get
		// values after that you need to access the map directly.
		"d": "1234",
	}
	for key, val := range arr {
		if ds.Value(key) != val {
			t.Error(fmt.Sprintf("! Expected '%s', got '%s'", val, ds.Value(key)))
		}
	}

}

func TestFormValuesDatasource(t *testing.T) {
	method := "POST"
	urlString := "http://example.com/form/test"
	body := strings.NewReader("name=Inigo+Montoya&fingers=6")

	request, err := http.NewRequest(method, urlString, body)

	// Canary
	if err != nil {
		t.Error("! Error constructing a request.", err)
	}

	// For POST requests with a body from a form the Content-Type header needs
	// to be set or the form body isn't processed.
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	ds := new(FormValuesDatasource).Init(request)

	if ds.Value("name").(string) != "Inigo Montoya" {
		t.Error("! Prepare to die.")
	}

	if ds.Value("fingers") != "6" {
		t.Error("! Expected six fingers, but got less.")
	}
}

func TestPathDatasource(t *testing.T) {
	ds := new(PathDatasource).Init("/foo/bar")
	if ds.Value("0") != "foo" {
		t.Error("! Expected value 0 to be 'foo'. Got ", ds.Value("0"))
	}

	if ds.Value("1") != "bar" {
		t.Error("! Expected value 1 to be 'bar'. Got ", ds.Value("1"))
	}

	ds = new(PathDatasource).Init("POST /foo/bar")
	if ds.Value("0") != "foo" {
		t.Error("! Expected value 0 to be 'foo'. Got ", ds.Value("0"))
	}

	if ds.Value("1") != "bar" {
		t.Error("! Expected value 1 to be 'bar'. Got ", ds.Value("1"))
	}

	ds = new(PathDatasource).Init("POST /foo/bar/baz/a/b/c/d/e/f")
	if ds.Value("0") != "foo" {
		t.Error("! Expected value 0 to be 'foo'. Got ", ds.Value("0"))
	}

	if ds.Value("1") != "bar" {
		t.Error("! Expected value 1 to be 'bar'. Got ", ds.Value("1"))
	}
}
