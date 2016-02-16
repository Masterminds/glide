// +build windows

package msg

// The color codes here are for compatibility with how Colors are used. Windows
// colors have not been implemented yet. See https://github.com/Masterminds/glide/issues/158
// for more detail.
const (
	Blue   = ""
	Red    = ""
	Green  = ""
	Yellow = ""
	Cyan   = ""
	Pink   = ""
)

// Color on windows returns no color. See
// https://github.com/Masterminds/glide/issues/158 if you want to help.
func (m *Messenger) Color(code, msg string) string {
	return msg
}
