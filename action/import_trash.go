package action

import (
	"github.com/Masterminds/glide/msg"
	"github.com/Masterminds/glide/trash"
)

// ImporTrash imports a Trash vendor file.
func ImporTrash(dest string) {
	base := "."
	config := EnsureConfig()
	if !trash.Has(base) {
		msg.Die("No Trash vendor file found.")
	}
	deps, err := trash.Parse(base)
	if err != nil {
		msg.Die("Failed to extract Trash vendor file: %s", err)
	}
	appendImports(deps, config)
	writeConfigToFileOrStdout(config, dest)
}
