// Package cache provides an interface for interfacing with the Glide local cache
//
// Glide has a local cache of metadata and repositories similar to the GOPATH.
// To store the cache Glide creates a .glide directory with a cache subdirectory.
// This is usually in the users home directory unless there is no accessible
// home directory in which case the .glide directory is in the root of the
// repository.
//
// To get the cache location use the `cache.Location()` function. This will
// return the proper base location in your environment.
//
// Within the cache directory there are two subdirectories. They are the src
// and info directories. The src directory contains version control checkouts
// of the packages. The info direcory contains metadata. The metadata maps to
// the RepoInfo struct. Both stores are happed to keys.
//
// Using the `cache.Key()` function you can get a key for a repo. Pass in a
// location such as `https://github.com/foo/bar` or `git@example.com:foo.git`
// and a key will be returned that can be used for caching operations.
//
// Note, the caching is based on repo rather than package. This is important
// for a couple reasons.
//
// 1. Forks or package replacements are supported in Glide. Where a different
//    repo maps to a package.
// 2. Permissions enable different access. For example `https://example.com/foo.git`
//    and `git@example.com:foo.git` may have access to different branches or tags.
package cache

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/glide/msg"
	gpath "github.com/Masterminds/glide/path"
)

// Enabled sets if the cache is globally enabled. Defaults to true.
var Enabled = true

// ErrCacheDisabled is returned with the cache is disabled.
var ErrCacheDisabled = errors.New("Cache disabled")

var isSetup bool

var setupMutex sync.Mutex

// Setup creates the cache location.
func Setup() error {
	setupMutex.Lock()
	defer setupMutex.Unlock()

	if isSetup {
		return nil
	}
	msg.Debug("Setting up the cache directory")
	pths := []string{
		"cache",
		filepath.Join("cache", "src"),
		filepath.Join("cache", "info"),
	}

	for _, l := range pths {
		err := os.MkdirAll(filepath.Join(gpath.Home(), l), 0755)
		if err != nil {
			return err
		}
	}

	isSetup = true
	return nil
}

// SetupReset resets if setup has been completed. The next time setup is run
// it will attempt a full setup.
func SetupReset() {
	isSetup = false
}

// Location returns the location of the cache.
func Location() (string, error) {
	p := filepath.Join(gpath.Home(), "cache")
	err := Setup()

	return p, err
}

// scpSyntaxRe matches the SCP-like addresses used to access repos over SSH.
var scpSyntaxRe = regexp.MustCompile(`^([a-zA-Z0-9_]+)@([a-zA-Z0-9._-]+):(.*)$`)

// Key generates a cache key based on a url or scp string. The key is file
// system safe.
func Key(repo string) (string, error) {

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

// RepoInfo holds information about a repo.
type RepoInfo struct {
	DefaultBranch string `json:"default-branch"`
	LastUpdate    string `json:"last-update"`
}

// SaveRepoData stores data about a repo in the Glide cache
func SaveRepoData(key string, data RepoInfo) error {
	if !Enabled {
		return ErrCacheDisabled
	}
	location, err := Location()
	if err != nil {
		return err
	}
	data.LastUpdate = time.Now().String()
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	pp := filepath.Join(location, "info")
	err = os.MkdirAll(pp, 0755)
	if err != nil {
		return err
	}

	p := filepath.Join(pp, key+".json")
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(d)
	return err
}

// RepoData retrieves cached information about a repo.
func RepoData(key string) (*RepoInfo, error) {
	if !Enabled {
		return &RepoInfo{}, ErrCacheDisabled
	}
	location, err := Location()
	if err != nil {
		return &RepoInfo{}, err
	}
	c := &RepoInfo{}
	p := filepath.Join(location, "info", key+".json")
	f, err := ioutil.ReadFile(p)
	if err != nil {
		return &RepoInfo{}, err
	}
	err = json.Unmarshal(f, c)
	if err != nil {
		return &RepoInfo{}, err
	}
	return c, nil
}

var lockSync sync.Mutex

var lockData = make(map[string]*sync.Mutex)

// Lock locks a particular key name
func Lock(name string) {
	lockSync.Lock()
	m, ok := lockData[name]
	if !ok {
		m = &sync.Mutex{}
		lockData[name] = m
	}
	lockSync.Unlock()
	msg.Debug("Locking %s", name)
	m.Lock()
}

// Unlock unlocks a particular key name
func Unlock(name string) {
	msg.Debug("Unlocking %s", name)
	lockSync.Lock()
	if m, ok := lockData[name]; ok {
		m.Unlock()
	}

	lockSync.Unlock()
}
