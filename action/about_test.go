package action

import (
	"bytes"
	"testing"

	"github.com/Masterminds/glide/msg"
)

func TestAbout(t *testing.T) {
	var buf bytes.Buffer
	old := msg.Default.Stdout
	msg.Default.Stdout = &buf
	About()

	if buf.Len() < len(aboutMessage) {
		t.Errorf("expected this to match aboutMessage: %q", buf.String())
	}

	msg.Default.Stdout = old
}
