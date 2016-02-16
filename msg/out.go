// +build !windows

package msg

import "fmt"

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
func (m *Messenger) Color(code, msg string) string {
	if m.NoColor {
		return msg
	}
	return fmt.Sprintf("\033[%sm%s\033[m", code, msg)
}
