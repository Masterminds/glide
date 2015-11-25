// Copyright 2013 Masterminds

package cookoo

import (
	"bytes"
	"regexp"
	"testing"
)

func TestLogMessage(t *testing.T) {
	registry, router, context := Cookoo()

	logger := new(bytes.Buffer)
	context.AddLogger("test", logger)

	registry.Route("test", "Testing.").
		Does(LogMessage, "logmsg").
		Using("msg").WithDefault("a test").
		Using("level").WithDefault("error")

	e := router.HandleRequest("test", context, false)

	if e != nil {
		t.Error("! Unexpected error.")
	}

	line := logger.String()
	line = line[0 : len(line)-1]
	pattern := "a test$"
	matched, err := regexp.MatchString(pattern, line)
	if err != nil {
		t.Fatal("! Regex Pattern did not compile:", err)
	}
	if !matched {
		t.Errorf("! Message was not logged to first test logger: %q", line)
	}
}

func TestAddToContext(t *testing.T) {
	registry, router, context := Cookoo()

	registry.Route("test", "Testing.").
		Does(AddToContext, "addtocontext").
		Using("foo").WithDefault("baz").
		Using("bar").WithDefault("qux")

	e := router.HandleRequest("test", context, false)

	if e != nil {
		t.Error("! Unexpected error.")
	}

	if context.Get("foo", "") != "baz" {
		t.Error("! foo was not added to the context with a value of baz.")
	}
	if context.Get("bar", "") != "qux" {
		t.Error("! bar was not added to the context with a value of qux.")
	}
}

func TestForwardTo(t *testing.T) {
	registry, router, context := Cookoo()

	registry.
		// Routes to test a basic working forward.
		Route("test", "Testing.").
		Does(ForwardTo, "forwardto").
		Using("route").WithDefault("test2").
		Route("test2", "A second test route").
		Does(AddToContext, "addtocontext").
		Using("bar").WithDefault("qux").

		// Route to test forward when no route supplied.
		Route("nope", "A test route to fatal error").
		Does(ForwardTo, "forwardto").
		Does(AddToContext, "addtocontext").
		Using("foo").WithDefault("baz").

		// Route to test with an ignored route.
		Route("test3", "Testing.").
		Does(ForwardTo, "forwardto").
		Using("route").WithDefault("nope").
		Using("ignoreRoutes").WithDefault([]string{"nope"}).
		Does(AddToContext, "addtocontext").
		Using("bar").WithDefault("qux")

	// Testing the case the required route is not supplied.
	e := router.HandleRequest("nope", context, false)
	if e == nil {
		t.Error("! Expected error executing nope")
	}
	v := context.Get("foo", nil)
	if v != nil {
		t.Error("! Expected AddToContext to not be executed.")
	}

	// Test a successful jump.
	e = router.HandleRequest("test", context, false)
	if e != nil {
		t.Error("! Unexpected error.")
	}
	v = context.Get("bar", nil)
	if v != "qux" {
		t.Error("! Expected test route to forward to test to adding bar to the context.")
	}

	// Test an ignored route
	e = router.HandleRequest("test3", context, false)
	if e != nil {
		t.Error("! Unexpected error.")
	}
	v = context.Get("bar", nil)
	if v != "qux" {
		t.Error("! Expected test3 route to forward to test to adding bar to the context.")
	}
}
