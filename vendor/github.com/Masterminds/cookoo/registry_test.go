package cookoo

import (
	"testing"
	//	"registry"
	"fmt"
)

type FooType struct {
	test int
}

func FakeCommand(cxt Context, params Params) (interface{}, Interrupt) {
	fmt.Println("Got here")

	var ret bool = true

	p := &ret

	return p, nil
}

func AnotherCommand(cxt Context, params *Params) (interface{}, Interrupt) {
	//ret := func() bool {return true;}
	ret := new(FooType)
	ret.test = 5

	return ret, nil
}

type Dossier struct {
	Name    string
	Age     int
	Alias   string `coo:"AKA"`
	RealAge int    `coo:"-"`
}

func (a *Dossier) Run(c Context) (interface{}, Interrupt) {
	return true, nil
}

func TestBasicRoute(t *testing.T) {
	reg := new(Registry)
	reg.Init()

	reg.Route("foo", "A test route")
	reg.Does(AnotherCommand, "fakeCommand").Using("param").WithDefault("value")

	// Now do something to test.
	routes := reg.Routes()

	if len(routes) != 1 {
		t.Error("! Expected one route.")
	}

	rspec := routes["foo"]

	if rspec.name != "foo" {
		t.Error("! Expected route to be named 'foo'")
	}
	if rspec.description != "A test route" {
		t.Error("! Expected description to be 'A test route'")
	}

	if len(rspec.commands) != 1 {
		t.Error("! Expected exactly one command. Found ", len(rspec.commands))
	}

	cmd := rspec.commands[0]
	if "fakeCommand" != cmd.name {
		t.Error("! Expected to find fakeCommand command.")
	}

	if len(cmd.parameters) != 1 {
		t.Error("! Expected exactly one paramter. Found ", len(cmd.parameters))
	}

	pspec := cmd.parameters[0]
	if pspec.name != "param" {
		t.Error("! Expected the first param to be 'param'")
	}

	if pspec.defaultValue != "value" {
		t.Error("! Expected the value to be 'value'")
	}
	fakeCxt := new(ExecutionContext)
	fakeParams := NewParamsWithValues(map[string]interface{}{"foo": "bar", "baz": 2})
	rr, err := cmd.command(fakeCxt, fakeParams)

	if err != nil {
		t.Error("! Expected no errors.")
	}

	cRet := rr.(*FooType)

	if cRet.test != 5 {
		t.Error("! Expected 'test' to be 5")
	}

}

func TestRouteIncludes(t *testing.T) {
	reg := new(Registry)
	reg.Init()

	reg.Route("foo", "A test route").
		Does(AnotherCommand, "fakeCommand").
		Using("param").WithDefault("foo").
		Route("bar", "Another test route").
		Does(AnotherCommand, "fakeCommand2").
		Using("param").WithDefault("bar").
		Includes("foo").
		Does(AnotherCommand, "fakeCommand3").
		Using("param").WithDefault("baz")

	expecting := []string{"fakeCommand2", "fakeCommand", "fakeCommand3"}
	spec, ok := reg.RouteSpec("bar")
	if !ok {
		t.Error("! Expected to find a route named 'bar'")
	}
	for i, k := range expecting {
		if k != spec.commands[i].name {
			t.Error(fmt.Sprintf("Expecting %s at position %d; got %s", k, i, spec.commands[i].name))
		}
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("! Failed to panic when including commands for a route that does not exist.")
		}
	}()
	reg2 := new(Registry)
	reg2.Init()

	reg2.Route("foo", "A test route").
		Includes("bar")
}

func TestRouteSpec(t *testing.T) {
	reg := new(Registry)
	reg.Init()

	reg.Route("foo", "A test route").
		Does(AnotherCommand, "fakeCommand").
		Using("param").WithDefault("value").
		Using("something").WithDefault(NewContext())

	spec, ok := reg.RouteSpec("foo")

	if !ok {
		t.Error("! Expected to find a route named 'foo'")
	}

	if spec.name != "foo" {
		t.Error("! Expected a spec named 'foo'")
	}

	param := spec.commands[0].parameters[1]
	if v, ok := param.defaultValue.(Context); !ok {
		t.Error("! Expected an execution context.")
	} else {
		// Canary
		v.Put("test", "test")
	}
}

func TestRouteNames(t *testing.T) {
	reg := new(Registry)
	reg.Init()
	reg.Route("one", "A route").Does(AnotherCommand, "fake")
	reg.Route("two", "A route").Does(AnotherCommand, "fake")
	reg.Route("three", "A route").Does(AnotherCommand, "fake")
	reg.Route("four", "A route").Does(AnotherCommand, "fake")
	reg.Route("five", "A route").Does(AnotherCommand, "fake")

	names := reg.RouteNames()

	if len(names) != 5 {
		t.Error("! Expected five routes, found ", len(names))
	}

	expecting := []string{"one", "two", "three", "four", "five"}
	for i, k := range expecting {
		if k != names[i] {
			t.Error(fmt.Sprintf("Expecting %s at position %d; got %s", k, i, names[i]))
		}
	}

}

func TestAddRoutes(t *testing.T) {
	reg := NewRegistry()
	err := reg.AddRoutes(
		Route{
			Name: "@boot",
			Does: Tasks{
				Cmd{Name: "startup"},
				Cmd{Name: "ready"},
			},
		},
		Route{
			Name: "Foo",
			Help: "Bar",
			Does: Tasks{
				Cmd{
					Name: "cmd1",
					Fn:   AnotherCommand,
					Using: []Param{
						{"test", 1, "cxt:test"},
						{"test2", "foo", "cxt:test2"},
					},
				},
				Cmd{
					Name: "cmd2",
				},
				// This will include all of the commands on route @boot.
				Include{"@boot"},
				Cmd{
					Name: "cmd3",
				},
			},
		},
		Route{
			Name: "Another",
			Does: Tasks{
				Cmd{
					Name: "cmd3",
				},
				Cmd{
					Name: "cmd4",
				},
				CmdDef{
					Name: "obj",
					Def:  &Dossier{},
				},
			},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	spec, ok := reg.RouteSpec("Foo")
	if !ok {
		t.Error("Expected to find route Foo.")
	}
	if spec.Name() != "Foo" {
		t.Errorf("Expected route named 'Foo', got '%s'", spec.Name())
	}
	if spec.Description() != "Bar" {
		t.Errorf("Expected description 'Bar', got '%s'", spec.Description())
	}

	if len(spec.commands) != 5 {
		t.Errorf("Expected exactly 5 commands, got %d", len(spec.commands))
	}

	// Check that each command is in the order we expect
	order := []string{"cmd1", "cmd2", "startup", "ready", "cmd3"}
	for i, c := range spec.commands {
		if c.name != order[i] {
			t.Errorf("Expected commnd %d to be %s, got %s", i, order[i], c.name)
		}
	}

	cmd := spec.commands[0]
	if cmd.name != "cmd1" {
		t.Errorf("Expected command named cmd1, got %s", cmd.name)
	}
	if cmd.command == nil {
		t.Error("Expected AnotherCommand.")
	}
	if len(cmd.parameters) != 2 {
		t.Errorf("Expected 2 paramters, got %d", len(cmd.parameters))
	}
	param := cmd.parameters[0]
	if param.name != "test" {
		t.Errorf("Expected param named 'test', got '%s'", param.name)
	}
	if param.defaultValue.(int) != 1 {
		t.Error("Expected int 1.")
	}
	if param.from != "cxt:test" {
		t.Error("Expected cxt:test, got %s", param.from)
	}
}
