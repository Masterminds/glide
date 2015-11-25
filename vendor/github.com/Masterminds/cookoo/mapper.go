package cookoo

import (
	//"fmt"
	"reflect"
	"strings"
)

// CommandDefinition defines what a command objects looks like.
//
// Fields on a command struct may be annotated with tags.
// A tag looks like this:
// 	type Foo struct {
// 		Field         string `coo:"myfield"`
// 		AnotherField  string `coo:"-" // - means skip this field.
// 		LastField     string `coo:"lastfield,cxt"
// 	}
//
// A tag has the format `coo:"NAME,MODIFIERS"`.
//
// When NAME is present, the reader will look in the params list for a field
// my that name. If no tag is present or the name field is blank, the field's
// name will be used instead.
//
// As with the JSON tags, if the field has the name "-" it will be skipped. No
// value will be set for it.
//
// The following modifiers are defined:
// 	- cxt: Get the value from the context. (EXPERIMENTAL)
// 	- ds: Get the value from the Datasources list. (EXPERIMENTAL)
// 	- param: Get the value from Params (default). (EXPERIMENTAL)
//
// When more than one is present (coo:"foo,cxt,ds") they will be tried in order.
// In this case, first the Context will be checked for foo, then the datasources.
//
// When no such modifier is present, the value is gotten only from the params list.

// CommandDefinition describes the fields that should be attached to a Command.
//
// CommandDefinitions should be composed of public fields and a Run method.
//
// 	type Person struct {
// 		Name string
// 		Age int
//		IgnoreMe string `coo:"-"`
// 	}
type CommandDefinition interface {
	// Run provides the same functionality as a Command function, but with
	// the added benefit of having all of the fields present on the struct. For that
	// reason, there is no Params attached.
	Run(c Context) (interface{}, Interrupt)
}

// Map merges params into a CommandDefinition and returns a Command Definition.
func Map(c Context, p *Params, d CommandDefinition) (CommandDefinition, Interrupt) {

	// For each field on the command definition, attempt to give it a value
	// from the params.
	v := reflect.Indirect(reflect.ValueOf(d))
	t := v.Type()
	count := t.NumField()

	// Make a new instance of the d
	def := reflect.New(t)

	// Iterate through all the fields on the struct. We want to
	// find the name of the field, or perhaps the tag's name.
	for i := 0; i < count; i++ {
		f := t.Field(i)

		fieldName := f.Name
		tag := []string{fieldName}

		// If there are tags, we need to parse them.
		if len(f.Tag) != 0 {
			//fmt.Printf("Need to parse tags %s.\n", f.Tag)
			t := f.Tag.Get("coo")
			if len(tag) > 0 {
				tag = parseTag(fieldName, t)
			}
		}

		fv := reflect.Indirect(def).FieldByName(f.Name)

		if err := populate(f, fv, c, p, tag); err != nil {
			return nil, err
		}
	}

	iface := def.Interface()
	return iface.(CommandDefinition), nil
}

// parseTag parses the contents of a coo tag.
func parseTag(fieldName, tag string) []string {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return []string{fieldName}
	}
	return parts
}

func populate(f reflect.StructField, v reflect.Value, c Context, p *Params, tag []string) error {

	//fmt.Printf("Reflecting on %s for %s\n", f.Name, tag[0])

	// Ignore fields with name "-".
	if tag[0] == "-" {
		return nil
	}

	//if !v.CanSet() {
	//	return fmt.Errorf("Field %s cannot be set.", f.Name)
	//}

	// Get only from Params.
	if len(tag) == 1 {
		val, ok := p.Has(tag[0])
		if ok {
			v.Set(reflect.ValueOf(val))
		}
		return nil
	}

	for i := 1; i < len(tag); i++ {
		switch tag[i] {
		case "cxt":
			val, ok := c.Has(tag[0])
			if ok {
				v.Set(reflect.ValueOf(val))
				return nil
			}
		case "param":
			val, ok := p.Has(tag[0])
			if ok {
				v.Set(reflect.ValueOf(val))
				return nil
			}
		case "ds":
			val, ok := c.HasDatasource(tag[0])
			if ok {
				v.Set(reflect.ValueOf(val))
				return nil
			}
		}

	}

	return nil
}
