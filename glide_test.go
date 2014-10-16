package main

import (
	"testing"

	"github.com/Masterminds/cookoo"
)

func TestCommandsNonEmpty(t *testing.T) {
	_, router, ctx := cookoo.Cookoo()
	commands := commands(ctx, router)
	if len(commands) == 0 {
		t.Fail()
	}
}
