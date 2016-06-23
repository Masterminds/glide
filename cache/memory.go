package cache

import (
	"sync"

	"github.com/Masterminds/glide/msg"
	"github.com/Masterminds/semver"
)

// Provide an in memory cache of imported project information.

var defaultMemCache = newMemCache()

// MemPut put a version into the in memory cache for a name.
// This will silently ignore non-semver and make sure the latest
// is stored.
func MemPut(name, version string) {
	defaultMemCache.put(name, version)
}

// MemTouched returns true if the cache was touched for a name.
func MemTouched(name string) bool {
	return defaultMemCache.touched(name)
}

// MemTouch notes if a name has been looked at.
func MemTouch(name string) {
	defaultMemCache.touch(name)
}

// MemLatest returns the latest, that is most recent, semver release. This
// may be a blank string if no put value
func MemLatest(name string) string {
	return defaultMemCache.getLatest(name)
}

// MemSetCurrent is used to set the current version in use.
func MemSetCurrent(name, version string) {
	defaultMemCache.setCurrent(name, version)
}

// MemCurrent is used to get the current version in use.
func MemCurrent(name string) string {
	return defaultMemCache.current(name)
}

// An in memory cache.
type memCache struct {
	sync.RWMutex
	latest   map[string]string
	t        map[string]bool
	versions map[string][]string
	c        map[string]string
}

func newMemCache() *memCache {
	return &memCache{
		latest:   make(map[string]string),
		t:        make(map[string]bool),
		versions: make(map[string][]string),
		c:        make(map[string]string),
	}
}

func (m *memCache) setCurrent(name, version string) {
	m.Lock()
	defer m.Unlock()

	if m.c[name] == "" {
		m.c[name] = version
	} else {
		// If we already have a version try to see if the new or old one is
		// semver and use that one.
		_, err := semver.NewVersion(m.c[name])
		if err != nil {
			_, err2 := semver.NewVersion(version)
			if err2 == nil {
				m.c[name] = version
			}
		}
	}
}

func (m *memCache) current(name string) string {
	m.RLock()
	defer m.RUnlock()
	return m.c[name]
}

func (m *memCache) put(name, version string) {
	m.Lock()
	defer m.Unlock()
	m.t[name] = true
	sv, err := semver.NewVersion(version)
	if err != nil {
		msg.Debug("Ignoring %s version %s: %s", name, version, err)
		return
	}

	latest, found := m.latest[name]
	if found {
		lv, err := semver.NewVersion(latest)
		if err == nil {
			if sv.GreaterThan(lv) {
				m.latest[name] = version
			}
		}
	} else {
		m.latest[name] = version
	}

	found = false
	for _, v := range m.versions[name] {
		if v == version {
			found = true
		}
	}
	if !found {
		m.versions[name] = append(m.versions[name], version)
	}
}

func (m *memCache) touch(name string) {
	m.Lock()
	defer m.Unlock()
	m.t[name] = true
}

func (m *memCache) touched(name string) bool {
	m.RLock()
	defer m.RUnlock()
	return m.t[name]
}

func (m *memCache) getLatest(name string) string {
	m.RLock()
	defer m.RUnlock()
	return m.latest[name]
}
