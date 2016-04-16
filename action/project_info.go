package action

import (
	"bytes"

	"github.com/Masterminds/glide/msg"
)

func Info(format string) {
	conf := EnsureConfig()
	var buffer bytes.Buffer
	varInit := false
	for _, var_format := range format {
		if varInit {
			switch var_format {
			case 'n':
				buffer.WriteString(conf.Name)
			case 'd':
				buffer.WriteString(conf.Description)
			case 'h':
				buffer.WriteString(conf.Home)
			case 'l':
				buffer.WriteString(conf.License)
			default:
				msg.Die("Invalid format %s", string(var_format))
			}
		} else {
			switch var_format {
			case '%':
				varInit = true
				continue
			default:
				buffer.WriteString(string(var_format))
			}
		}
		varInit = false
	}
	msg.Puts(buffer.String())
}
