package action

import (
	"testing"

	"github.com/Masterminds/glide/msg"
)

func TestNoVendor(t *testing.T) {
	msg.Default.PanicOnDie = true
	NoVendor("../testdata/nv", false, false)
}
