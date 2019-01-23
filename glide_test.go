package glide

import (
	"testing"
)

func TestCommandsNonEmpty(t *testing.T) {
	commands := Commands()
	if len(commands) == 0 {
		t.Fail()
	}
}
