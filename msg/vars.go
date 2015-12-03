package msg

import (
	"fmt"
	"os"
)

// Quiet, if true, suppresses chatty levels, like Info.
var Quiet = false

// IsDebugging, if true, shows verbose levels, like Debug.
var IsDebugging = false

// NoColor, if true, will not use color in the output.
var NoColor = false

// Stdout is the location where this prints output.
var Stdout = os.Stdout

// Stderr is the location where this prints logs.
var Stderr = os.Stderr

// Puts formats a message and then prints to Stdout.
//
// It does not prefix the message, does not color it, or otherwise decorate it.
//
// It does add a line feed.
func Puts(msg string, args ...interface{}) {
	fmt.Fprintf(Stdout, msg, args...)
	fmt.Fprintln(Stdout)
}
