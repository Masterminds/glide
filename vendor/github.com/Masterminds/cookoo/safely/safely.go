/* Safely is a package for providing safety wrappers around commonly used features.

You might have at one point done this:

	go myFunc()

But what happens if myFunc panics? It will terminate the program with a panic.

Safely provides a panic-trapping goroutine runner:

	safely.Go(myFunc)

If `myFunc` panics, Safely will capture the panic and log the error message.
*/
package safely

import (
	"log"
)

// GoDoer is a function that `safely.Go` can execute.
type GoDoer func()

// SafelyLogger is used to log messages when a failure occurs
type SafelyLogger interface {
	Printf(string, ...interface{})
}

// Captures the log portion of a cookoo.Context.
type Logger interface {
	Logf(string, string, ...interface{})
}

// Go executes a function as a goroutine, but recovers from any panics.
//
// Normally, if a GoRoutine panics, it will stop execution on the current
// program, which ain't always good.
//
// safely.Go handles this by trapping the panic and writing it to the default
// logger.
//
// To use your own logger, use safely.GoLog.
func Go(todo GoDoer) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				// Seems like there should be a way to get the default logger
				// and then pass this into GoLog.
				log.Printf("Panic in safely.Go: %s", err)
			}
		}()
		todo()
	}()
}

// GoDo runs a Goroutine, traps panics, and logs panics to a safely.Logger.
//
// Example:
//
// 	_, _, cxt := cookoo.Cookoo()
// 	safely.GoDo(cxt, func(){})
//
//
func GoDo(cxt Logger, todo GoDoer) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				cxt.Logf("error", "Panic in safely.Go: %s", err)
			}
		}()
		todo()
	}()
}


// GoLog executes a function as a goroutine, but traps any panics.
//
// If a panic is encountered, it is logged to the given logger.
func GoLog(logger SafelyLogger, todo GoDoer) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Printf("Panic in safely.Go: %s", err)
			}
		}()
		todo()
	}()
}
