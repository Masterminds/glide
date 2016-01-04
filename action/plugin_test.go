package action

import (
	"os"
	"testing"

	"github.com/Masterminds/glide/msg"
)

func TestPlugin(t *testing.T) {
	wd, _ := os.Getwd()
	os.Chdir("../testdata/plugin")
	msg.Default.PanicOnDie = true
	cmd := "hello"
	args := []string{"a", "b"}
	// FIXME: Trapping the panic is the nice thing to do.
	Plugin(cmd, args)
	os.Chdir(wd)
}
