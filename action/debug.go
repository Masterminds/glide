package action

import (
	"github.com/Masterminds/glide/msg"
)

// Debug sets the debugging flags across components.
func Debug(on bool) {
	msg.Default.IsDebugging = on

	if on == true {
		msg.Default.IsVerbose = on
	}
}

// Verbose sets the verbose flags across components.
func Verbose(on bool) {
	msg.Default.IsVerbose = on
}

// Quiet sets the quiet flags across components.
func Quiet(on bool) {
	msg.Default.Quiet = on
}

// NoColor sets the color flags.
func NoColor(on bool) {
	msg.Default.NoColor = on
}
