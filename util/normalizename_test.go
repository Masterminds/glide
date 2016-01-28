package util

import (
	"testing"
)

func TestNormalizeName(t *testing.T) {
	packages := map[string]string{
		"github.com/Masterminds/cookoo/web/io/foo": "github.com/Masterminds/cookoo",
		"golang.org/x/crypto/ssh":                  "golang.org/x/crypto",
		"incomplete/example":                       "incomplete/example",
		"net":                                      "net",
	}
	for start, expected := range packages {
		if finish, extra := NormalizeName(start); expected != finish {
			t.Errorf("Expected '%s', got '%s'", expected, finish)
		} else if start != finish && start != finish+"/"+extra {
			t.Errorf("Expected %s to end with %s", finish, extra)
		}
	}
}
