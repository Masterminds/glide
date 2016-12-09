package action

import (
	"github.com/Masterminds/glide/govendor"
	"github.com/Masterminds/glide/msg"
)

// ImportGovendor imports govendor dependencies into the present glide config
func ImportGovendor(dest string) {
	base := "."
	config := EnsureConfig()
	if !govendor.Has(base) {
		msg.Die("There is no govendor file to import.")
	}
	deps, err := govendor.Parse(base)
	if err != nil {
		msg.Die("Failed to extract govendor file: %s", err)
	}
	appendImports(deps, config)
	writeConfigToFileOrStdout(config, dest)
}
