// Borrowed from golang.org/x/build/cmd/fetchlogs/xdg.go
//
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license.

package cache

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// xdgCacheDir returns the XDG Base Directory Specification cache
// directory.
func xdgCacheDir() string {
	cache := os.Getenv("XDG_CACHE_HOME")
	if cache == "" {
		home := os.Getenv("HOME")
		if home == "" {
			u, err := user.Current()
			if err != nil {
				home = u.HomeDir
			}
		}
		// Not XDG but standard for OS X.
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library/Caches")
		}
		cache = filepath.Join(home, ".cache")
	}
	return cache
}
