package cookoo

import (
	"testing"
)

func TestParams(t *testing.T) {
	params := NewParamsWithValues(map[string]interface{}{
		"Test":  123,
		"Test2": "Hello",
		"Test3": NewContext(),
	})

	if v, ok := params.Has("Test"); !ok {
		t.Error("Expected to find 123, got NADA")
	} else if v != 123 {
		t.Error("! Expected 123, got ", v)
	}

	// A really lame validator.
	fn := func(value interface{}) bool {
		return true
	}

	// Test the validator.
	if v, ok := params.Validate("Test2", fn); !ok {
		t.Error("! Expected a valid string.")
	} else if v != "Hello" {
		t.Error("! Expected 'Hello', got ", v)
	}

	alwaysFails := func(value interface{}) bool {
		return false
	}
	// Test the validator.
	if _, ok := params.Validate("Test2", alwaysFails); ok {
		t.Error("! Expected a failed validation.")
	}

	if v, ok := params.Has("Test3"); !ok {
		t.Error("! Expected a context in Test3.")
	} else if _, ok = v.(Context); !ok {
		t.Error("! Expected the value to be a Context.")
	}

	if ok, missing := params.Requires("Test", "Test3"); !ok {
		t.Error("Expected to find params. Missing ", missing)
	}

	if ok, missing := params.Requires("Test", "Test4"); ok {
		t.Error("! Expected to be missing something", missing)
	} else if missing[0] != "Test4" {
		t.Error("! Expected to be missing Test4")
	}

	get := params.Get("Test2", "NOT IT")
	if get != "Hello" {
		t.Error("! Expected Hello, got ", get)
	}

	get = params.Get("NotThere", "YES")
	if get != "YES" {
		t.Error("! Expected YES, got ", get)
	}
}
