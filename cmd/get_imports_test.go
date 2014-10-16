package cmd

import (
	"testing"

	"github.com/Masterminds/cookoo"
)

func TestGetImportsEmptyConfig(t *testing.T) {
	_, _, c := cookoo.Cookoo()
	SilenceLogs(c)
	cfg := new(Config)
	p := cookoo.NewParamsWithValues(map[string]interface{}{"conf": cfg})
	res, it := GetImports(c, p)
	if it != nil {
		t.Errorf("Interrupt value non-nil")
	}
	bres, ok := res.(bool)
	if !ok || bres {
		t.Errorf("Result was non-bool or true: ok=%t bres=%t", ok, bres)
	}
}

func SilenceLogs(c cookoo.Context) {
	p := cookoo.NewParamsWithValues(map[string]interface{}{"quiet": true})
	BeQuiet(c, p)
}
