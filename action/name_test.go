package action

import (
	"bytes"
	"os"
	"testing"

	"github.com/Masterminds/glide/msg"
)

func TestName(t *testing.T) {
	var buf bytes.Buffer
	msg.Default.PanicOnDie = true
	ostdout := msg.Default.Stdout
	msg.Default.Stdout = &buf
	wd, _ := os.Getwd()
	if err := os.Chdir("../testdata/name"); err != nil {
		t.Errorf("Failed to change directory: %s", err)
	}
	Name()
	if buf.String() != "technosophos.com/x/foo\n" {
		t.Errorf("Unexpectedly got name %q", buf.String())
	}
	msg.Default.Stdout = ostdout
	os.Chdir(wd)
}
