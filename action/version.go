package action

import (
	"github.com/Masterminds/glide/msg"
)

func Version() {
	conf := EnsureConfig()
	msg.Puts(conf.Version)
}
