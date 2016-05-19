package repo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	//"github.com/Masterminds/glide/msg"
)

var cacheEnabled = true

var errCacheDisabled = errors.New("Cache disabled")

// EnsureCacheDir Creates the $HOME/.glide/cache directory (unless home is
// specified to be different) if it does not exist.
/*
func EnsureCacheDir(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	home := p.Get("home", "").(string)
	if home == "" {
		cacheEnabled = false
		msg.Warn("Unable to locate home directory")
		return false, nil
	}
	err := os.MkdirAll(filepath.Join(home, "cache", "info"), os.ModeDir|os.ModePerm)
	if err != nil {
		cacheEnabled = false
		Warn("Error creating Glide directory %s", home)
	}
	return false, nil
}
*/

// scpSyntaxRe matches the SCP-like addresses used to access repos over SSH.
var scpSyntaxRe = regexp.MustCompile(`^([a-zA-Z0-9_]+)@([a-zA-Z0-9._-]+):(.*)$`)

// Pass in a repo location and get a cache key from it.
func cacheCreateKey(repo string) (string, error) {

	var u *url.URL
	var err error
	var strip bool
	if m := scpSyntaxRe.FindStringSubmatch(repo); m != nil {
		// Match SCP-like syntax and convert it to a URL.
		// Eg, "git@github.com:user/repo" becomes
		// "ssh://git@github.com/user/repo".
		u = &url.URL{
			Scheme: "ssh",
			User:   url.User(m[1]),
			Host:   m[2],
			Path:   "/" + m[3],
		}
		strip = true
	} else {
		u, err = url.Parse(repo)
		if err != nil {
			return "", err
		}
	}

	if strip {
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

type cacheRepoInfo struct {
	DefaultBranch string `json:"default-branch"`
	LastUpdate    string `json:"last-update"`
}

func saveCacheRepoData(key string, data cacheRepoInfo, location string) error {
	if !cacheEnabled {
		return errCacheDisabled
	}
	data.LastUpdate = time.Now().String()
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	p := filepath.Join(location, "cache", "info", key+".json")
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(d)
	return err
}

func cacheRepoData(key, location string) (*cacheRepoInfo, error) {
	if !cacheEnabled {
		return &cacheRepoInfo{}, errCacheDisabled
	}
	c := &cacheRepoInfo{}
	p := filepath.Join(location, "cache", "info", key+".json")
	f, err := ioutil.ReadFile(p)
	if err != nil {
		return &cacheRepoInfo{}, err
	}
	err = json.Unmarshal(f, c)
	if err != nil {
		return &cacheRepoInfo{}, err
	}
	return c, nil
}
