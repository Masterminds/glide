package util

import (
	"testing"
)

func TestNormalizeName(t *testing.T) {
	packages := []struct {
		input string
		root  string
		extra string
	}{
		{
			input: "github.com/Masterminds/cookoo/web/io/foo",
			root:  "github.com/Masterminds/cookoo",
			extra: "web/io/foo",
		},
		{
			input: `github.com\Masterminds\cookoo\web\io\foo`,
			root:  "github.com/Masterminds/cookoo",
			extra: "web/io/foo",
		},
		{
			input: "golang.org/x/crypto/ssh",
			root:  "golang.org/x/crypto",
			extra: "ssh",
		},
		{
			input: "incomplete/example",
			root:  "incomplete/example",
			extra: "",
		},
		{
			input: "net",
			root:  "net",
			extra: "",
		},
	}
	for _, test := range packages {
		root, extra := NormalizeName(test.input)
		if root != test.root {
			t.Errorf("%s: Expected root '%s', got '%s'", test.input, test.root, root)
		}
		if extra != test.extra {
			t.Errorf("%s: Expected extra '%s', got '%s'", test.input, test.extra, extra)
		}
	}
}
