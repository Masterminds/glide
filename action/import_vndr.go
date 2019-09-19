package action

import (
	"github.com/Masterminds/glide/msg"
	"github.com/Masterminds/glide/vndr"
)

// ImportVNDR imports a vendor.conf file.
func ImportVNDR(dest string) {
	base := "."
	config := EnsureConfig()
	if !vndr.Has(base) {
		msg.Die("No VNDR data found.")
	}
	deps, err := vndr.Parse(base)
	if err != nil {
		msg.Die("Failed to extract VNDR file: %s", err)
	}
	appendImports(deps, config)
	writeConfigToFileOrStdout(config, dest)
}
