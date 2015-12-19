package action

import (
	"os"
	"testing"

	"github.com/Masterminds/glide/msg"
)

func TestPlugin(t *testing.T) {
	os.Chdir("../testdata/plugin")
	msg.PanicOnDie = true
	cmd := "hello"
	args := []string{"a", "b"}
	// FIXME: Trapping the panic is the nice thing to do.
	Plugin(cmd, args)
}
