package repo

import (
	"sync"
)

// UpdateTracker holds a list of all the packages that have been updated from
// an external source. This is a concurrency safe implementation.
type UpdateTracker struct {
	sync.RWMutex

	updated map[string]bool
}

// NewUpdateTracker creates a new instance of UpdateTracker ready for use.
func NewUpdateTracker() *UpdateTracker {
	u := &UpdateTracker{}
	u.updated = map[string]bool{}
	return u
}

// Add adds a name to the list of items being tracked.
func (u *UpdateTracker) Add(name string) {
	u.Lock()
	u.updated[name] = true
	u.Unlock()
}

// Check returns if an item is on the list or not.
func (u *UpdateTracker) Check(name string) bool {
	u.RLock()
	_, f := u.updated[name]
	u.RUnlock()
	return f
}

// Remove takes a package off the list
func (u *UpdateTracker) Remove(name string) {
	u.Lock()
	delete(u.updated, name)
	u.Unlock()
}
