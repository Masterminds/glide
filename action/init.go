package action

import (
	gpath "github.com/Masterminds/glide/path"
)

// Init initializes the action subsystem for handling one or more subesequent actions.
func Init(yaml, home string) {
	gpath.GlideFile = yaml
	gpath.SetHome(home)
}
