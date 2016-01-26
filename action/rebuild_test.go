package action

import (
	"os"
	"testing"

	"github.com/Masterminds/glide/msg"
)

func TestRebuild(t *testing.T) {
	msg.Default.PanicOnDie = true
	wd, _ := os.Getwd()
	if err := os.Chdir("../testdata/rebuild"); err != nil {
		t.Errorf("Could not change dir: %s (%s)", err, wd)
	}
	Rebuild()
	os.Chdir(wd)
}
