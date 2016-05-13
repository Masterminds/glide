// Package cache provides an interface for interfacing with the Glide local cache
package cache

import (
	"path/filepath"

	gpath "github.com/Masterminds/glide/path"
)

// Location returns the location of the cache.
func Location() string {
	return filepath.Join(gpath.Home(), "cache")
}
