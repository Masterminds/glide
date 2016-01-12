package fmt

import (
	"github.com/Masterminds/cookoo"
	"testing"
)

func TestSprintf(t *testing.T) {
	reg, router, cxt := cookoo.Cookoo();

	reg.Route("test", "Test").
	Does(Sprintf, "out").
	Using("format").WithDefault("%s %d").
	Using("0").WithDefault("Hello").
	Using("1").WithDefault(1)

	if err := router.HandleRequest("test", cxt, false); err != nil {
		t.Errorf("Failed route: %s", err)
	}

	if res := cxt.Get("out", "absolutely nothin"); res != "Hello 1" {
		t.Errorf("Expected 'Hello 1', got %s", res)
	}
}

func TestTemplate(t *testing.T) {
	reg, router, cxt := cookoo.Cookoo();

	reg.Route("test", "Test").
	Does(Template, "out").
	Using("template").WithDefault("{{.Hello}} {{.one}}").
	Using("Hello").WithDefault("Hello").
	Using("one").WithDefault(1)

	if err := router.HandleRequest("test", cxt, false); err != nil {
		t.Errorf("Failed route: %s", err)
	}

	if res := cxt.Get("out", "nada"); res != "Hello 1" {
		t.Errorf("Expected 'Hello 1', got %s", res)
	}

	reg.Route("test2", "Test 2").
	Does(cookoo.AddToContext, "_").Using("Foo").WithDefault("lambkin").
	Does(Template, "out2").
	Using("template.Context").WithDefault(true).
	Using("template").WithDefault("Hello {{.Cxt.Foo}}")

	if err := router.HandleRequest("test2", cxt, false); err != nil {
		t.Errorf("Failed route: %s", err)
	}

	if res := cxt.Get("out2", "nada"); res != "Hello lambkin" {
		t.Errorf("Expected 'Hello lambkin', got %s", res)
	}
}
