/* Package log contains logging helpers.

EXPERIMENTAL! This may change prior to the final release of this version!

This package provides syslog-style logging messages. See `log/syslog` in Go's
documentation. A notable difference, though, is that this also provides
formatter variants for all of the log levels.

This uses the Context.Log* functions beneath the hood, so any logger
configuration for that will also hold true for this.
*/
package log

import (
	"runtime"

	"github.com/Masterminds/cookoo"
)

// LogLevel describes the log levels.
type LogLevel uint8

// Log levels correspond to UNIX/syslog logging levels.
const (
	LogEmerg LogLevel = iota
	LogAlert
	LogCrit
	LogErr
	LogWarning
	LogNotice
	LogInfo
	LogDebug
)

// Labels for log levels.
const (
	LabelEmerg   = "[emergency] "
	LabelAlert   = "[alert] "
	LabelCrit    = "[critical] "
	LabelErr     = "[error] "
	LabelWarning = "[warning] "
	LabelNotice  = "[notice] "
	LabelInfo    = "[info] "
	LabelDebug   = "[debug] "
)

var Label = [8]string{
	LabelEmerg,
	LabelAlert,
	LabelCrit,
	LabelErr,
	LabelWarning,
	LabelNotice,
	LabelInfo,
	LabelDebug,
}

var Level LogLevel = LogDebug

// Debugging returns true if the level is set to allow debugging.
//
// Whether or not the log message is sent to the underlying logger is determined
// based on the Level. However, using checks like this can prevent doing
// costly debug computations just for the sake of logging.
//
// 	if Debugging() {
// 		// Do something expensive.
// 		costlyOperation()
//
// 		Debug(c, "msg")
// 	}
//
// Otherwise, this will write the message to the lower-level logger, which can
// then decide (presumably based on Level) what to do with the message.
func Debugging() bool {
	return Level >= LogDebug
}

// Emergf logs an emergency.
func Emergf(c cookoo.Context, msg string, args ...interface{}) {
	sendf(c, LogEmerg, msg, args...)
}

// Alertf logs an alert.
func Alertf(c cookoo.Context, msg string, args ...interface{}) {
	sendf(c, LogAlert, msg, args...)
}

// Critf logs a critical message.
func Critf(c cookoo.Context, msg string, args ...interface{}) {
	sendf(c, LogCrit, msg, args...)
}

// Errf logs an error message.
func Errf(c cookoo.Context, msg string, args ...interface{}) {
	sendf(c, LogErr, msg, args...)
}

// Warnf logs a warning.
func Warnf(c cookoo.Context, msg string, args ...interface{}) {
	sendf(c, LogWarning, msg, args...)
}

// Noticef logs a notice.
func Noticef(c cookoo.Context, msg string, args ...interface{}) {
	sendf(c, LogNotice, msg, args...)
}

// Infof logs an informational message.
func Infof(c cookoo.Context, msg string, args ...interface{}) {
	sendf(c, LogInfo, msg, args...)
}

// Debugf logs a debug message.
func Debugf(c cookoo.Context, msg string, args ...interface{}) {
	sendf(c, LogDebug, msg, args...)
}

// Emergf logs an emergency.
func Emerg(c cookoo.Context, args ...interface{}) {
	send(c, LogEmerg, args...)
}

// Alertf logs an alert.
func Alert(c cookoo.Context, args ...interface{}) {
	send(c, LogAlert, args...)
}

// Critf logs a critical message.
func Crit(c cookoo.Context, args ...interface{}) {
	send(c, LogCrit, args...)
}

// Errf logs an error message.
func Err(c cookoo.Context, args ...interface{}) {
	send(c, LogErr, args...)
}

// Warnf logs a warning.
func Warn(c cookoo.Context, args ...interface{}) {
	send(c, LogWarning, args...)
}

// Noticef logs a notice.
func Notice(c cookoo.Context, args ...interface{}) {
	send(c, LogNotice, args...)
}

// Infof logs an informational message.
func Info(c cookoo.Context, args ...interface{}) {
	send(c, LogInfo, args...)
}

// Debugf logs a debug message.
func Debug(c cookoo.Context, args ...interface{}) {
	send(c, LogDebug, args...)
}

// Stack dumps a stack trace to the log. It uses the LogDebug level.
//
// This limits the size of the returned stack to 4096 bytes.
func Stack(c cookoo.Context, msg string) {
	buff := make([]byte, 4096)
	runtime.Stack(buff, false)
	send(c, LogDebug, msg, string(buff))
}

func send(c cookoo.Context, l LogLevel, args ...interface{}) {
	if Level >= l {
		c.Log(Label[l], args...)
	}
}
func sendf(c cookoo.Context, l LogLevel, msg string, args ...interface{}) {
	if Level >= l {
		c.Logf(Label[l], msg, args...)
	}
}
