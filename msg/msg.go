package msg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// Messenger provides the underlying implementation that displays output to
// users.
type Messenger struct {
	sync.Mutex

	// Quiet, if true, suppresses chatty levels, like Info.
	Quiet bool

	// IsDebugging, if true, shows Debug.
	IsDebugging bool

	// IsVerbose, if true, shows detailed informational messages.
	IsVerbose bool

	// NoColor, if true, will not use color in the output.
	NoColor bool

	// Stdout is the location where this prints output.
	Stdout io.Writer

	// Stderr is the location where this prints logs.
	Stderr io.Writer

	// Stdin is the location where input is read.
	Stdin io.Reader

	// PanicOnDie if true Die() will panic instead of exiting.
	PanicOnDie bool

	// The default exit code to use when dyping
	ecode int

	// If an error was been sent.
	hasErrored bool
}

// NewMessenger creates a default Messenger to display output.
func NewMessenger() *Messenger {
	m := &Messenger{
		Quiet:       false,
		IsDebugging: false,
		IsVerbose:   false,
		NoColor:     false,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Stdin:       os.Stdin,
		PanicOnDie:  false,
		ecode:       1,
	}

	return m
}

// Default contains a default Messenger used by package level functions
var Default = NewMessenger()

// Info logs information
func (m *Messenger) Info(msg string, args ...interface{}) {
	if m.Quiet {
		return
	}
	prefix := m.Color(Green, "[INFO]\t")
	m.Msg(prefix+msg, args...)
}

// Info logs information using the Default Messenger
func Info(msg string, args ...interface{}) {
	Default.Info(msg, args...)
}

// Debug logs debug information
func (m *Messenger) Debug(msg string, args ...interface{}) {
	if m.Quiet || !m.IsDebugging {
		return
	}
	prefix := "[DEBUG]\t"
	m.Msg(prefix+msg, args...)
}

// Debug logs debug information using the Default Messenger
func Debug(msg string, args ...interface{}) {
	Default.Debug(msg, args...)
}

// Verbose logs detailed information
func (m *Messenger) Verbose(msg string, args ...interface{}) {
	if m.Quiet || !m.IsVerbose {
		return
	}
	m.Info(msg, args...)
}

// Verbose detailed information using the Default Messenger
func Verbose(msg string, args ...interface{}) {
	Default.Verbose(msg, args...)
}

// Warn logs a warning
func (m *Messenger) Warn(msg string, args ...interface{}) {
	prefix := m.Color(Yellow, "[WARN]\t")
	m.Msg(prefix+msg, args...)
}

// Warn logs a warning using the Default Messenger
func Warn(msg string, args ...interface{}) {
	Default.Warn(msg, args...)
}

// Err logs an error.
func (m *Messenger) Err(msg string, args ...interface{}) {
	prefix := m.Color(Red, "[ERROR]\t")
	m.Msg(prefix+msg, args...)
	m.hasErrored = true
}

// Err logs anderror using the Default Messenger
func Err(msg string, args ...interface{}) {
	Default.Err(msg, args...)
}

// Die prints an error message and immediately exits the application.
// If PanicOnDie is set to true a panic will occur instead of os.Exit being
// called.
func (m *Messenger) Die(msg string, args ...interface{}) {
	m.Err(msg, args...)
	if m.PanicOnDie {
		panic("trapped a Die() call")
	}
	os.Exit(m.ecode)
}

// Die prints an error message and immediately exits the application using the
// Default Messenger. If PanicOnDie is set to true a panic will occur instead of
// os.Exit being called.
func Die(msg string, args ...interface{}) {
	Default.Die(msg, args...)
}

// ExitCode sets the exit code used by Die.
//
// The default is 1.
//
// Returns the old error code.
func (m *Messenger) ExitCode(e int) int {
	m.Lock()
	old := m.ecode
	m.ecode = e
	m.Unlock()
	return old
}

// ExitCode sets the exit code used by Die using the Default Messenger.
//
// The default is 1.
//
// Returns the old error code.
func ExitCode(e int) int {
	return Default.ExitCode(e)
}

// Msg prints a message with optional arguments, that can be printed, of
// varying types.
func (m *Messenger) Msg(msg string, args ...interface{}) {
	// When operations in Glide are happening concurrently messaging needs to be
	// locked to avoid displaying one message in the middle of another one.
	m.Lock()
	defer m.Unlock()

	// Get rid of the annoying fact that messages need \n at the end, but do
	// it in a backward compatible way.
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	if len(args) == 0 {
		fmt.Fprint(m.Stderr, msg)
	} else {
		fmt.Fprintf(m.Stderr, msg, args...)
	}
}

// Msg prints a message with optional arguments, that can be printed, of
// varying types using the Default Messenger.
func Msg(msg string, args ...interface{}) {
	Default.Msg(msg, args...)
}

// Puts formats a message and then prints to Stdout.
//
// It does not prefix the message, does not color it, or otherwise decorate it.
//
// It does add a line feed.
func (m *Messenger) Puts(msg string, args ...interface{}) {
	// When operations in Glide are happening concurrently messaging needs to be
	// locked to avoid displaying one message in the middle of another one.
	m.Lock()
	defer m.Unlock()

	fmt.Fprintf(m.Stdout, msg, args...)
	fmt.Fprintln(m.Stdout)
}

// Puts formats a message and then prints to Stdout using the Default Messenger.
//
// It does not prefix the message, does not color it, or otherwise decorate it.
//
// It does add a line feed.
func Puts(msg string, args ...interface{}) {
	Default.Puts(msg, args...)
}

// Print prints exactly the string given.
//
// It prints to Stdout.
func (m *Messenger) Print(msg string) {
	// When operations in Glide are happening concurrently messaging needs to be
	// locked to avoid displaying one message in the middle of another one.
	m.Lock()
	defer m.Unlock()

	fmt.Fprint(m.Stdout, msg)
}

// Print prints exactly the string given using the Default Messenger.
//
// It prints to Stdout.
func Print(msg string) {
	Default.Print(msg)
}

// HasErrored returns if Error has been called.
//
// This is useful if you want to known if Error was called to exit with a
// non-zero exit code.
func (m *Messenger) HasErrored() bool {
	return m.hasErrored
}

// HasErrored returns if Error has been called on the Default Messenger.
//
// This is useful if you want to known if Error was called to exit with a
// non-zero exit code.
func HasErrored() bool {
	return Default.HasErrored()
}

// Color returns a string in a certain color if colors are enabled and
// available on that platform.
func Color(code, msg string) string {
	return Default.Color(code, msg)
}

// PromptUntil provides a prompt until one of the passed in strings has been
// entered and return is hit. Note, the comparisons are case insensitive meaning
// Y == y. The returned value is the one from the passed in options (same case).
func (m *Messenger) PromptUntil(opts []string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		for _, c := range opts {
			if strings.EqualFold(c, strings.TrimSpace(text)) {
				return c, nil
			}
		}
	}
}

// PromptUntil provides a prompt until one of the passed in strings has been
// entered and return is hit. Note, the comparisons are case insensitive meaning
// Y == y. The returned value is the one from the passed in options (same case).
// Uses the default setup.
func PromptUntil(opts []string) (string, error) {
	return Default.PromptUntil(opts)
}

// PromptUntilYorN provides a prompt until the user chooses yes or no. This is
// not case sensitive and they can input other options such as Y or N.
// In the response Yes is bool true and No is bool false.
func (m *Messenger) PromptUntilYorN() bool {
	res, err := m.PromptUntil([]string{"y", "yes", "n", "no"})
	if err != nil {
		m.Die("Error processing response: %s", err)
	}

	if res == "y" || res == "yes" {
		return true
	}

	return false
}

// PromptUntilYorN provides a prompt until the user chooses yes or no. This is
// not case sensitive and they can input other options such as Y or N.
// In the response Yes is bool true and No is bool false.
// Uses the default setup.
func PromptUntilYorN() bool {
	return Default.PromptUntilYorN()
}
