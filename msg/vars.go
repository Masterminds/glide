package msg

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Quiet, if true, suppresses chatty levels, like Info.
var Quiet = false

// IsDebugging, if true, shows verbose levels, like Debug.
var IsDebugging = false

// NoColor, if true, will not use color in the output.
var NoColor = false

// Stdout is the location where this prints output.
var Stdout io.Writer = os.Stdout

// Stderr is the location where this prints logs.
var Stderr io.Writer = os.Stderr

// If this is true, Die() will panic instead of exiting.
var PanicOnDie = false

var ecode = 1

var elock sync.Mutex

// Puts formats a message and then prints to Stdout.
//
// It does not prefix the message, does not color it, or otherwise decorate it.
//
// It does add a line feed.
func Puts(msg string, args ...interface{}) {
	fmt.Fprintf(Stdout, msg, args...)
	fmt.Fprintln(Stdout)
}

func Die(msg string, args ...interface{}) {
	Error(msg, args...)
	if PanicOnDie {
		panic("trapped a Die() call")
	}
	os.Exit(ecode)
}

// ExitCode sets the exit code used by Die.
//
// The default is 1.
//
// Returns the old error code.
func ExitCode(e int) int {
	elock.Lock()
	old := ecode
	ecode = e
	elock.Unlock()
	return old
}
