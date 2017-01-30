package action

import (
	"os"
	"runtime"
	"testing"

	"github.com/Masterminds/glide/msg"
)

func TestPlugin(t *testing.T) {
	wd, _ := os.Getwd()
	os.Chdir("../testdata/plugin")
	msg.Default.PanicOnDie = true
	var cmd string

	// Windows scripts for testing (batch) are different from shells scripts.
	// Making sure the plugin works in both bases.
	if runtime.GOOS == "windows" {
		cmd = "hello-win"
	} else {
		cmd = "hello"
	}
	args := []string{"a", "b"}
	// FIXME: Trapping the panic is the nice thing to do.
	Plugin(cmd, args)
	os.Chdir(wd)
}
