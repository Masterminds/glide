# Safely: Do scary things safely in Go.

This is a small utility library for doing things safely.

## Go Routine Wrappers

Don't let a panic in a `goroutine` take down your entire program.

```go
import "github.com/Masterminds/safely"

func main() {
  safely.Go(func() { panic("Oops!") })
}
```
