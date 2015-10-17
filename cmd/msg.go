// +build !windows

package cmd

import (
	"fmt"
	"os"
	"strings"
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

// Color returns a string in a certain color. The first argument is a string
// containing the color code or a constant from the table above mapped to a code.
//
// The following will print the string "Foo" in yellow:
//     fmt.Print(Color(Yellow, "Foo"))
func Color(code, msg string) string {
	return fmt.Sprintf("\033[%sm%s\033[m", code, msg)
}

// Info logs information
func Info(msg string, args ...interface{}) {
	if Quiet {
		return
	}
	fmt.Print(Color(Green, "[INFO] "))
	Msg(msg, args...)
}

// Debug logs debug information
func Debug(msg string, args ...interface{}) {
	if Quiet || !IsDebugging {
		return
	}
	fmt.Print("[DEBUG] ")
	Msg(msg, args...)
}

// Warn logs a warning
func Warn(msg string, args ...interface{}) {
	fmt.Fprint(os.Stderr, Color(Yellow, "[WARN] "))
	ErrMsg(msg, args...)
}

// Error logs and error.
func Error(msg string, args ...interface{}) {
	fmt.Fprint(os.Stderr, Color(Red, "[ERROR] "))
	ErrMsg(msg, args...)
}

// ErrMsg sends a message to Stderr
func ErrMsg(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg)
		return
	}
	fmt.Fprintf(os.Stderr, msg, args...)

	// Get rid of the annoying fact that messages need \n at the end, but do
	// it in a backward compatible way.
	if !strings.HasSuffix(msg, "\n") {
		fmt.Fprintln(os.Stderr)
	}
}

// Msg prints a message with optional arguments, that can be printed, of
// varying types.
func Msg(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Print(msg)
		return
	}
	fmt.Printf(msg, args...)

	// Get rid of the annoying fact that messages need \n at the end, but do
	// it in a backward compatible way.
	if !strings.HasSuffix(msg, "\n") {
		fmt.Println()
	}
}
