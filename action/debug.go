package action

import (
	"github.com/Masterminds/glide/msg"
)

// Debug sets the debugging flags across components.
func Debug(on bool) {
	msg.IsDebugging = on
}

// Quiet sets the quiet flags across components.
func Quiet(on bool) {
	msg.Quiet = on
}

// NoColor sets the color flags.
func NoColor(on bool) {
	msg.NoColor = on
}
