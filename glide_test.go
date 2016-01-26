package main

import (
	"testing"
)

func TestCommandsNonEmpty(t *testing.T) {
	commands := commands()
	if len(commands) == 0 {
		t.Fail()
	}
}
