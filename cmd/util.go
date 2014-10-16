package cmd

import (
	"fmt"
	"os"

	"github.com/Masterminds/cookoo"
)

// Quiet, when set to true, can suppress Info and Debug messages.
var Quiet = false

// These contanstants map to color codes for shell scripts making them
// human readable.
const (
	Blue    = "0;34"
	Red     = "0;31"
	BoldRed = "1;31"
	Yellow  = "0;33"
	Cyan    = "0;36"
	Pink    = "1;35"
)

// Color returns a string in a certain color. The first argument is a string
// containing the color code or a constant from the table above mapped to a code.
//
// The following will print the string "Foo" in yellow:
//     fmt.Print(Color(Yellow, "Foo"))
func Color(code, msg string) string {
	return fmt.Sprintf("\033[%sm%s\033[m", code, msg)
}

// BeQuiet supresses Info and Debug messages.
func BeQuiet(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	Quiet = p.Get("quiet", false).(bool)
	return Quiet, nil
}

// Info logs information
func Info(msg string, args ...interface{}) {
	if Quiet {
		return
	}
	fmt.Print(Color(Yellow, "[INFO] "))
	Msg(msg, args...)
}

// Debug logs debug information
func Debug(msg string, args ...interface{}) {
	if Quiet {
		return
	}
	fmt.Print("[DEBUG] ")
	Msg(msg, args...)
}

// Warn logs a warning
func Warn(msg string, args ...interface{}) {
	fmt.Fprint(os.Stderr, Color(Red, "[WARN] "))
	ErrMsg(msg, args...)
}

// Error logs and error.
func Error(msg string, args ...interface{}) {
	fmt.Fprint(os.Stderr, Color(BoldRed, "[ERROR] "))
	ErrMsg(msg, args...)
}

// ErrMsg sends a message to Stderr
func ErrMsg(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg)
		return
	}
	fmt.Fprintf(os.Stderr, msg, args...)
}

// Msg prints a message with optional arguments, that can be printed, of
// varying types.
func Msg(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Print(msg)
		return
	}
	fmt.Printf(msg, args...)
}
