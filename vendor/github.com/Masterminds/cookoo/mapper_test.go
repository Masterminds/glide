package cookoo

import (
	"testing"
)

func TestMap(t *testing.T) {
	cxt := NewContext()
	params := NewParamsWithValues(map[string]interface{}{
		"funty":          1,
		"FloatField":     float32(2.3),
		"StrField":       "hello",
		"BoolField":      true,
		"SliceField":     []int{1, 2, 3},
		"MapField":       map[string]bool{"true": true},
		"StructField":    basic{true},
		"StructPtrField": &basic{false},
	})
	s := &mystruct{}

	def, err := Map(cxt, params, s)
	if err != nil {
		t.Errorf("Failed: %s", err)
	}

	res := def.(*mystruct)

	if res.IntField != 1 {
		t.Errorf("Expected 1, got %d", res.IntField)
	}

	if res.FloatField != 2.3 {
		t.Errorf("Expected 2.3, got %f", res.FloatField)
	}

	if res.StrField != "hello" {
		t.Errorf("Expected hello, got %s", res.StrField)
	}

	if !res.BoolField {
		t.Errorf("BoolField is false")
	}

	if len(res.SliceField) != 3 {
		t.Errorf("Expected slice of 2.")
	}
	if !res.MapField["true"] {
		t.Errorf("Expected true:true.")
	}

	if !res.StructField.hai {
		t.Errorf("expected basic to have true")
	}

	if res.StructPtrField.hai {
		t.Errorf("expected *basic to have false.")
	}

}

func TestMappedRoute(t *testing.T) {
	reg, router, cxt := Cookoo()
	reg.AddRoute(Route{
		Name: "test",
		Does: Tasks{
			CmdDef{
				Name: "cmdDef",
				Def:  &mystruct{},
				Using: []Param{
					{Name: "funty", DefaultValue: 5},
					{Name: "StrField", DefaultValue: "Batty"},
				},
			},
		},
	})

	if err := router.HandleRequest("test", cxt, true); err != nil {
		t.Error(err)
	}

	cmdDef := cxt.Get("cmdDef", nil).(*mystruct)
	if cmdDef.IntField != 5 {
		t.Errorf("Expected 5, got %d", cmdDef.IntField)
	}
	if cmdDef.StrField != "Batty" {
		t.Errorf("Expected Batty, got %s", cmdDef.StrField)
	}

}

type mystruct struct {
	IntField       int `coo:"funty"`
	FloatField     float32
	StrField       string
	BoolField      bool
	SliceField     []int
	MapField       map[string]bool
	StructField    basic
	StructPtrField *basic
}

func (m *mystruct) Run(c Context) (interface{}, Interrupt) {
	return m, nil
}

type basic struct {
	hai bool
}
