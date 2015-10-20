// +build windows

package cmd

import (
	"fmt"
	"os"
	"strings"
)

// Info logs information
func Info(msg string, args ...interface{}) {
	if Quiet {
		return
	}
	fmt.Print("[INFO] ")
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
	fmt.Fprint(os.Stderr, "[WARN] ")
	ErrMsg(msg, args...)
}

// Error logs and error.
func Error(msg string, args ...interface{}) {
	fmt.Fprint(os.Stderr, "[ERROR] ")
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

	// Get rid of the annoying fact that messages need \n at the end, but do
	// it in a backward compatible way.
	if !strings.HasSuffix(msg, "\n") {
		fmt.Println("")
	}
}
