package action

import (
	"github.com/Masterminds/glide/msg"
)

func Debug(on bool) {
	msg.IsDebugging = on
}

func Quiet(on bool) {
	msg.Quiet = on
}

func NoColor(on bool) {
	msg.NoColor = on
}
