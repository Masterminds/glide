package cookoo

import "testing"

type testDs struct {
	val string
}
func (d *testDs) Value(key string) interface{} {
	return d.val
}

func TestGettableDS (t *testing.T) {

	ds := &testDs{"hello"}
	gds := GettableDS(ds)

	if "hello" != gds.Get("foo", "bar").(string) {
		t.Error("Expected hello.")
	}
}
func TestGettableContext (t *testing.T) {

	c := NewContext()
	c.Put("foo", "hello")
	gcx := GettableCxt(c)

	if "hello" != gcx.Get("foo", "bar").(string) {
		t.Error("Expected hello.")
	}
}

func TestGetters (t *testing.T) {

	p := NewParamsWithValues(map[string]interface{} {
		"bool": true,
		"string": "hello",
		"int64": int64(-1234567890),
		"uint64": uint64(1234567890),
		"float64": float64(0.1234),
	})

	if !GetBool("bool", false, p) {
		t.Error("Expected true")
	}
	if "hello" != GetString("string", "boo", p) {
		t.Error("Expected hello.")
	}

	if int64(-1234567890) != GetInt64("int64", 0, p) {
		t.Error("Expected -1234567890")
	}

	if uint64(1234567890) != GetUint64("uint64", 0, p) {
		t.Error("Expected 1234567890")
	}

	if float64(0.1234) != GetFloat64("float64", 0, p) {
		t.Error("Expected 0.1234")
	}

	if v, ok := HasBool("bool", p); !ok {
		t.Errorf("Expected to find bool")
	} else if !v {
		t.Errorf("Expected book to be true")
	}

	if v, ok := HasString("string", p); !ok {
		t.Error("Expected to find string")
	} else if v != "hello" {
		t.Errorf("Expected to get the string hello")
	}
}

func TestGetFromFirst(t *testing.T) {
	ds := &testDs{"hello"}
	gds := GettableDS(ds)
	c := NewContext()
	c.Put("test", "world")
	gcxt := GettableCxt(c)

	first, foundIn := GetFromFirst("test", "foo", gcxt, gds)
	if "world" != first {
		t.Errorf("Expected 'world', but got '%s'", first)
	}
	if _, ok := foundIn.(*gettableContext); !ok {
		t.Error("Expected to find in context.")
	}

	second, _ := GetFromFirst("bar", "foo", gcxt, gds)
	if "hello" != second {
		t.Errorf("Expected hello, got %s", second)
	}

	third, defSrc := GetFromFirst("bar", "foo", gcxt)
	if "foo" != third {
		t.Errorf("Expected foo, got %s", third)
	}

	if GetString("***", "***", defSrc) != "foo" {
		t.Errorf("Expected default datasource to always return foo.")
	}
}

