package action

import (
	"os"
	"path"
	"strings"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/msg"
)

// CacheClear clears the Glide cache
func CacheClear() {
	l := cache.Location()

	err := os.RemoveAll(l)
	if err != nil {
		msg.Die("Unable to clear the cache: %s", err)
	}

	cache.SetupReset()
	cache.Setup()

	msg.Info("Glide cache has been cleared.")
}

// CachePath prints the path to a cached dependency's sources.
func CachePath(importPath string) {
	for _, d := range EnsureConfig().Imports {
		if strings.EqualFold(importPath, d.Name) {
			k, err := cache.Key(d.Remote())
			if err != nil {
				msg.Die("Could not obtain cache key", err)
			}
			msg.Puts(path.Join(cache.Location(), "src", k))
			return
		}
	}
}
