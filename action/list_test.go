package action

import (
	"bytes"
	"testing"

	"github.com/Masterminds/glide/msg"
)

func TestList(t *testing.T) {
	var buf bytes.Buffer
	old := msg.Default.Stdout
	msg.Default.PanicOnDie = true
	msg.Default.Stdout = &buf
	List("../", false)
	if buf.Len() < 5 {
		t.Error("Expected some data to be found.")
	}
	// TODO: We should capture and test output.
	msg.Default.Stdout = old
}
