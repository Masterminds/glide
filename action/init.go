package action

import (
	"strings"
	"path/filepath"
	"os"
	"go/build"

	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

const _vendorSuffix = string(filepath.Separator) + "_vendor"

// Init initializes the action subsystem for handling one or more subesequent actions.
func Init(yaml, home string) {
	gpath.GlideFile = yaml
	gpath.HomeDir = home
}

func CheckGoVendor() {
	supportGoVendor := EnsureGoVendor()
	if !supportGoVendor {
		// if there is a *_vendorSuffix GOPATH, keep work with it
		gopath := filepath.SplitList(build.Default.GOPATH)
		var _vendor string
		for _, p := range gopath {
			if strings.HasSuffix(filepath.Clean(p), _vendorSuffix) {
				_vendor = p
				break
			}
		}
		if len(_vendor) == 0 {
			msg.Warn("no *%s GOPATH found.", _vendorSuffix)
			os.Exit(1)
		} else {
			msg.Info("found *%s GOPATH: %v", _vendorSuffix, _vendor)
			gpath.UseGoVendor = false
			gpath.GoPathVendorDir = filepath.Join(_vendor, "src")
		}
	}
}
