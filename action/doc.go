// Package action provides implementations for every Glide command.
//
// This is not a general-purpose library. It is the main flow controller for Glide.
//
// The main glide package acts as a Facade, with this package providing the
// implementation. This package should know nothing of the command line flags or
// runtime characteristics. However, this package is allowed to control the flow
// of the application, including termination. So actions may call `msg.Die()` to
// immediately stop execution of the program.
//
// In general, actions are not required to function as library functions, nor as
// concurrency-safe functions.
package action
