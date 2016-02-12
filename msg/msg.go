package msg

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/Masterminds/kitt/progress"
)

// Messanger provides the underlying implementation that displays output to
// users.
type Messanger struct {
	sync.Mutex

	// Quiet, if true, suppresses chatty levels, like Info.
	Quiet bool

	// IsDebugging, if true, shows verbose levels, like Debug.
	IsDebugging bool

	// NoColor, if true, will not use color in the output.
	NoColor bool

	// Stdout is the location where this prints output.
	Stdout io.Writer

	// Stderr is the location where this prints logs.
	Stderr io.Writer

	// PanicOnDie if true Die() will panic instead of exiting.
	PanicOnDie bool

	// InProgress indicates whether the Messanger is currently in a progress meter.
	InProgress bool

	// The default exit code to use when dyping
	ecode int

	// If an error was been sent.
	hasErrored bool

	meter ProgressMeter
}

type ProgressMeter interface {
	Start(string)
	Message(string)
	Done(string)
}

// NewMessanger creates a default Messanger to display output.
func NewMessanger() *Messanger {
	//var mtr ProgressMeter = (nilMeter)(0)
	mtr := progress.NewIndicator()
	mtr.Frames = Cylon
	mtr.Interval = CylonInterval
	m := &Messanger{
		Quiet:       false,
		IsDebugging: false,
		NoColor:     false,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		PanicOnDie:  false,
		ecode:       1,
		meter:       mtr,
		InProgress:  false,
	}

	return m
}

// Default contains a default messanger used by package level functions
var Default = NewMessanger()

// Info logs information
func (m *Messanger) Info(msg string, args ...interface{}) {
	if m.Quiet {
		return
	}
	prefix := m.Color(Green, "[INFO] ")
	m.Msg(prefix+msg, args...)
}

// Info logs information using the Default Messanger
func Info(msg string, args ...interface{}) {
	if Default.InProgress {
		Default.meter.Message(fmt.Sprintf(msg, args...))
		return
	}
	Default.Info(msg, args...)
}

// Debug logs debug information
func (m *Messanger) Debug(msg string, args ...interface{}) {
	if m.Quiet || !m.IsDebugging {
		return
	}
	prefix := "[DEBUG] "
	Msg(prefix+msg, args...)
}

// Debug logs debug information using the Default Messanger
func Debug(msg string, args ...interface{}) {
	Default.Debug(msg, args...)
}

// Warn logs a warning
func (m *Messanger) Warn(msg string, args ...interface{}) {
	prefix := m.Color(Yellow, "[WARN] ")
	m.Msg(prefix+msg, args...)
}

// Warn logs a warning using the Default Messanger
func Warn(msg string, args ...interface{}) {
	Default.Warn(msg, args...)
}

// Error logs and error.
func (m *Messanger) Error(msg string, args ...interface{}) {
	prefix := m.Color(Red, "[ERROR] ")
	m.Msg(prefix+msg, args...)
	m.hasErrored = true
}

// Error logs and error using the Default Messanger
func Error(msg string, args ...interface{}) {
	Default.Error(msg, args...)
}

// Die prints an error message and immediately exits the application.
// If PanicOnDie is set to true a panic will occur instead of os.Exit being
// called.
func (m *Messanger) Die(msg string, args ...interface{}) {
	m.Error(msg, args...)
	if m.PanicOnDie {
		panic("trapped a Die() call")
	}
	os.Exit(m.ecode)
}

// Die prints an error message and immediately exits the application using the
// Default Messanger. If PanicOnDie is set to true a panic will occur instead of
// os.Exit being called.
func Die(msg string, args ...interface{}) {
	Default.Die(msg, args...)
}

// ExitCode sets the exit code used by Die.
//
// The default is 1.
//
// Returns the old error code.
func (m *Messanger) ExitCode(e int) int {
	m.Lock()
	old := m.ecode
	m.ecode = e
	m.Unlock()
	return old
}

// ExitCode sets the exit code used by Die using the Default Messanger.
//
// The default is 1.
//
// Returns the old error code.
func ExitCode(e int) int {
	return Default.ExitCode(e)
}

// Msg prints a message with optional arguments, that can be printed, of
// varying types.
func (m *Messanger) Msg(msg string, args ...interface{}) {
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
// varying types using the Default Messanger.
func Msg(msg string, args ...interface{}) {
	Default.Msg(msg, args...)
}

// Puts formats a message and then prints to Stdout.
//
// It does not prefix the message, does not color it, or otherwise decorate it.
//
// It does add a line feed.
func (m *Messanger) Puts(msg string, args ...interface{}) {
	// When operations in Glide are happening concurrently messaging needs to be
	// locked to avoid displaying one message in the middle of another one.
	m.Lock()
	defer m.Unlock()

	fmt.Fprintf(m.Stdout, msg, args...)
	fmt.Fprintln(m.Stdout)
}

// Puts formats a message and then prints to Stdout using the Default Messanger.
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
func (m *Messanger) Print(msg string) {
	// When operations in Glide are happening concurrently messaging needs to be
	// locked to avoid displaying one message in the middle of another one.
	m.Lock()
	defer m.Unlock()

	fmt.Fprint(m.Stdout, msg)
}

// Print prints exactly the string given using the Default Messanger.
//
// It prints to Stdout.
func Print(msg string) {
	Default.Print(msg)
}

// HasErrored returns if Error has been called.
//
// This is useful if you want to known if Error was called to exit with a
// non-zero exit code.
func (m *Messanger) HasErrored() bool {
	return m.hasErrored
}

// HasErrored returns if Error has been called on the Default Messanger.
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
