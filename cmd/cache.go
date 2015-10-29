package cmd

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/cookoo"
)

// EnsureCacheDir Creates the $HOME/.glide/cache directory (unless home is
// specified to be different) if it does not exist.
func EnsureCacheDir(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	home := p.Get("home", "").(string)
	if home == "" {
		return nil, errors.New("No home directory set to create")
	}

	return os.MkdirAll(filepath.Join(home, "cache"), os.ModeDir|os.ModePerm), nil
}

// Pass in a repo location and get a cache key from it.
func cacheCreateKey(repo string) (string, error) {

	// A url needs a scheme. A git repo such as
	// git@github.com:Masterminds/cookoo.git reworked to the url parser.
	c := strings.Contains(repo, "://")
	if !c {
		repo = "ssh://" + repo
	}

	u, err := url.Parse(repo)
	if err != nil {
		return "", err
	}

	if !c {
		u.Scheme = ""
	}

	var key string
	if u.Scheme != "" {
		key = u.Scheme + "-"
	}
	if u.User != nil && u.User.Username() != "" {
		key = key + u.User.Username() + "-"
	}
	key = key + u.Host
	if u.Path != "" {
		key = key + strings.Replace(u.Path, "/", "-", -1)
	}

	key = strings.Replace(key, ":", "-", -1)

	return key, nil
}
