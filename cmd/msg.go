// +build !windows

package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// These contanstants map to color codes for shell scripts making them
// human readable.
const (
	Blue   = "0;34"
	Red    = "0;31"
	Green  = "0;32"
	Yellow = "0;33"
	Cyan   = "0;36"
	Pink   = "1;35"
)

var outputLock sync.Mutex

// Color returns a string in a certain color. The first argument is a string
// containing the color code or a constant from the table above mapped to a code.
//
// The following will print the string "Foo" in yellow:
//     fmt.Print(Color(Yellow, "Foo"))
func Color(code, msg string) string {
	if NoColor {
		return msg
	}
	return fmt.Sprintf("\033[%sm%s\033[m", code, msg)
}

// Info logs information
func Info(msg string, args ...interface{}) {
	if Quiet {
		return
	}
	i := fmt.Sprint(Color(Green, "[INFO] "))
	Msg(i+msg, args...)
}

// Debug logs debug information
func Debug(msg string, args ...interface{}) {
	if Quiet || !IsDebugging {
		return
	}
	i := fmt.Sprint("[DEBUG] ")
	Msg(i+msg, args...)
}

// Warn logs a warning
func Warn(msg string, args ...interface{}) {
	i := fmt.Sprint(Color(Yellow, "[WARN] "))
	ErrMsg(i+msg, args...)
}

// Error logs and error.
func Error(msg string, args ...interface{}) {
	i := fmt.Sprint(Color(Red, "[ERROR] "))
	ErrMsg(i+msg, args...)
}

// ErrMsg sends a message to Stderr
func ErrMsg(msg string, args ...interface{}) {
	outputLock.Lock()
	defer outputLock.Unlock()

	// If messages don't have a newline on the end we add one.
	e := ""
	if !strings.HasSuffix(msg, "\n") {
		e = "\n"
	}
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg+e)
	} else {
		fmt.Fprintf(os.Stderr, msg+e, args...)
	}
}

// Msg prints a message with optional arguments, that can be printed, of
// varying types.
func Msg(msg string, args ...interface{}) {
	outputLock.Lock()
	defer outputLock.Unlock()

	// If messages don't have a newline on the end we add one.
	e := ""
	if !strings.HasSuffix(msg, "\n") {
		e = "\n"
	}
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg+e)
	} else {
		fmt.Fprintf(os.Stderr, msg+e, args...)
	}
}
