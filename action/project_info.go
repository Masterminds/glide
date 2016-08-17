package action

import (
	"bytes"

	"github.com/Masterminds/glide/msg"
)

// Info prints information about a project based on a passed in format.
func Info(format string) {
	conf := EnsureConfig()
	var buffer bytes.Buffer
	varInit := false
	for _, varfmt := range format {
		if varInit {
			switch varfmt {
			case 'n':
				buffer.WriteString(conf.Name)
			case 'd':
				buffer.WriteString(conf.Description)
			case 'h':
				buffer.WriteString(conf.Home)
			case 'l':
				buffer.WriteString(conf.License)
			default:
				msg.Die("Invalid format %s", string(varfmt))
			}
		} else {
			switch varfmt {
			case '%':
				varInit = true
				continue
			default:
				buffer.WriteString(string(varfmt))
			}
		}
		varInit = false
	}
	msg.Puts(buffer.String())
}
