package action

import (
	"github.com/Masterminds/glide/msg"
)

// Name prints the name of the package, according to the glide.yaml file.
func Name(yamlpath string) {
	conf := EnsureConfig(yamlpath)
	msg.Puts(conf.Name)
}
