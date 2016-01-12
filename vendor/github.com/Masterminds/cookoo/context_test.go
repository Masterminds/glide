// Copyright 2013 Masterminds

// This package provides the execution context for a Cookoo request.
package cookoo

import (
	"bytes"
	"log"
	"reflect"
	"regexp"
	"runtime"
	"testing"
)

// An example datasource as can add to our store.
type ExampleDatasource struct {
	name string
}

// A simple equal function.
func equal(t *testing.T, a interface{}, b interface{}) {
	result := reflect.DeepEqual(a, b)
	if !result {
		_, file, line, _ := runtime.Caller(1)
		t.Errorf("Failed equals in %s:%d", file, line)
	}
}

func TestDatasource(t *testing.T) {
	foo := new(ExampleDatasource)
	foo.name = "bar"

	cxt := NewContext()

	cxt.AddDatasource("foo", foo)

	foo2 := cxt.Datasource("foo").(*ExampleDatasource)

	equal(t, foo, foo2)
	equal(t, "bar", foo2.name)

	cxt.RemoveDatasource("foo")

	equal(t, nil, cxt.Datasource("foo"))
}

func TestAddGet(t *testing.T) {
	cxt := NewContext()

	cxt.Put("test1", 42)
	cxt.Put("test2", "Geronimo!")
	cxt.Put("test3", func() string { return "Hello" })

	// Test Get
	equal(t, 42, cxt.Get("test1", nil))
	equal(t, "Geronimo!", cxt.Get("test2", nil))

	// Test has
	val, ok := cxt.Has("test1")
	if !ok {
		t.Error("! Failed to get 'test1'")
	}
	equal(t, 42, cxt.Get("test1", nil))

	_, ok = cxt.Has("test999")
	if ok {
		t.Error("! Unexpected result for 'test999'")
	}

	val, ok = cxt.Has("test3")
	fn := val.(func() string)
	if ok {
		equal(t, "Hello", fn())
	} else {
		t.Error("! Expected a function.")
	}

	m := cxt.AsMap()
	if m["test1"] != 42 {
		t.Error("! Error retrieving context as a map.")
	}
}

type LameStruct struct {
	stuff []string
}

func TestCopy(t *testing.T) {
	lame := new(LameStruct)
	lame.stuff = []string{"O", "Hai"}
	c := NewContext()
	c.Put("a", lame)
	c.Put("b", "This is the song that never ends")

	foo := new(ExampleDatasource)
	foo.name = "bar"
	c.AddDatasource("foo", foo)

	c2 := c.Copy()

	c.Put("c", 1234)

	if c.Len() != 3 {
		t.Error("! Canary failed. c should be 3")
	}

	if c2.Len() != 2 {
		t.Error("! c2 should be 2.")
	}

	c.Put("b", "FOO")
	if c2.Get("b", nil) == "FOO" {
		t.Error("! b should not have changed in C2.")
	}

	lame.stuff[1] = "Noes"

	v1 := c2.Get("a", nil).(*LameStruct)
	if v1.stuff[1] != "Noes" {
		t.Error("! Expected shallow copy of array. Got ", v1)
	}

	d2 := new(ExampleDatasource)
	d2.name = "bar"
	c.AddDatasource("d2", d2)
	_, found := c2.HasDatasource("d2")
	if found == true {
		t.Error("! Datasource inadvertently copied from one context to another.")
	}
}

func TestLogging(t *testing.T) {
	logger := new(bytes.Buffer)
	c := NewContext()
	c.AddLogger("test", logger)
	foo, found := c.Logger("test")

	if found == false {
		t.Error("! Logger not found.")
	}

	if foo != logger {
		t.Error("! Loggers do not match.")
	}

	logger2 := new(bytes.Buffer)
	c.AddLogger("test2", logger2)

	c.(*ExecutionContext).SkipLogPrefix("tinker", "bell")

	log.Print("a test")
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

	logger2.Reset()
	c.Log("foo", "this is a test")
	line = logger2.String()
	line = line[0 : len(line)-1]
	pattern = "^foo.*this is a test$"
	matched, err = regexp.MatchString(pattern, line)
	if err != nil {
		t.Fatal("! Regex Pattern did not compile:", err)
	}
	if !matched {
		t.Errorf("! Log to second logger did not happen correctly: %q", line)
	}

	logger.Reset()
	c.Logf("tinker", "ignore")
	c.Logf("bar", "foo %d baz", 2)
	c.Log("bell", "ignore")
	line = logger.String()
	line = line[0 : len(line)-1]
	pattern = "^bar.*foo 2 baz$"
	matched, err = regexp.MatchString(pattern, line)
	if err != nil {
		t.Fatal("! pattern did not compile:", err)
	}
	if !matched {
		t.Errorf("! Logf to first logger did now happen correctly: %q", line)
	}

	c.RemoveLogger("test")
	_, found = c.Logger("test")

	if found == true {
		t.Error("! Logger found but should have been removed.")
	}

}
