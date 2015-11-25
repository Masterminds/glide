package convert

import (
	"github.com/Masterminds/cookoo"
	"testing"
)

func TestAtoi(t *testing.T) {
	reg, router, c := cookoo.Cookoo()
	reg.Route("test", "Test convert.").Does(Atoi, "i").Using("str").From("cxt:a")

	c.Put("a", "100")
	e := router.HandleRequest("test", c, false)

	if e != nil {
		t.Errorf("! Failed during HandleRequest: %s", e)
		return
	}
	i, ok := c.Has("i")
	if !ok {
		t.Error("! Expected to find 'a' in context, but it was missing.")
	}
	if i.(int) != 100 {
		t.Errorf("! Expected '100' to be converted to 100. Got %d", i)
	}
}
