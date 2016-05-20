package action

import (
	"os"

	"github.com/Masterminds/glide/cache"
	"github.com/Masterminds/glide/msg"
)

// CacheClear clears the Glide cache
func CacheClear() {
	l, err := cache.Location()
	if err != nil {
		msg.Die("Unable to clear the cache: %s", err)
	}

	err = os.RemoveAll(l)
	if err != nil {
		msg.Die("Unable to clear the cache: %s", err)
	}

	cache.SetupReset()
	err = cache.Setup()
	if err != nil {
		msg.Die("Unable to clear the cache: %s", err)
	}

	msg.Info("Glide cache has been cleared.")
}
