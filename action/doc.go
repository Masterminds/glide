/* Package action provides implementations for every Glide command.

The main glide package acts as a Facade, with this package providing the
implementation. This package should know nothing of the command line flags or
runtime characteristics. However, this package is allowed to indicate that a
particular action should be aborted. So actions may call `msg.Die()` to
immediately stop execution of the program.

In general, actions are not required to function as library functions, nor as
concurrency-safe functions.
*/
package action
