package cookoo

import (
	"sync"
	"io"
)

// SyncContext wraps a context, syncronizing access to it.
//
// This uses a read/write mutex which allows multiple reads at a time, but
// locks both reading and writing for writes.
//
// To avoid really nefarious bugs, the same mutex locks context values and
// datasource values (since there is no guarantee that one is not backed by
// the other).
func SyncContext(cxt Context) Context {
	var mutex sync.RWMutex
	return &synchronizedContext{mutex: mutex, cxt: cxt}
}

type synchronizedContext struct {
	mutex sync.RWMutex
	cxt Context
}

// Add is deprecated. Use Put instead.
func (s *synchronizedContext) Add(key string, val ContextValue) {
	s.Put(key, val)
}

// Put locks the context and then inserts the key/value pair.
func (s *synchronizedContext) Put (key string, val ContextValue) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cxt.Put(key, val)
}

// Get pulls a readlock, and then returns the associated value (or default
// if no suitable value is found).
func (s *synchronizedContext) Get(key string, def interface{}) ContextValue {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cxt.Get(key, def)
}

// Has locks the context, and retrieves value and OK.
//
// If the key is not found in the context, the ok is set to false.
func (s *synchronizedContext) Has(key string) (ContextValue, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cxt.Has(key)
}

// Datasource read-locks the context, and then fetches the named datasource.
func (s *synchronizedContext) Datasource(key string) Datasource {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cxt.Datasource(key)
}
// Datasources read-locks and then returns a copy of a map of datasources.
//
// WARNING: As with all datasource operations, it is up to the underlying
// datasource to handle concurrency issues. This does not deep copy the
// datasources.
func (s *synchronizedContext) Datasources() map[string]Datasource {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// XXX: Is there any reason to copy the map before returning?
	return s.cxt.Datasources()
}
// HasDatasource read-locks the context and then fetches the named datasource.
//
// If the datasource is not found, the ok flag will be set to false. This
// allows you to differentiate between a missing datasource and a nil datasource.
func (s *synchronizedContext) HasDatasource(key string) (Datasource, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cxt.HasDatasource(key)
}
// AddDatasource locks the context and then adds a datasource.
func (s *synchronizedContext) AddDatasource(key string, ds Datasource) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cxt.AddDatasource(key, ds)
}
// RemoveDatasource locks the context and then removes a datasource.
func (s *synchronizedContext) RemoveDatasource(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cxt.RemoveDatasource(key)
}
// Len read-locks the context and then finds the length.
func (s *synchronizedContext) Len() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cxt.Len()
}

// Copy makes a shallow copy of the underlying context, and then wraps it in a
// new synchronizer.
//
// Use this with care. Because the copy is shallow, you can expose race
// conditions when two contexts access the same underlying data. You may
// want to make your own deep copy of the context. (see `Context.AsMap()`)
func (s *synchronizedContext) Copy() Context {
	return SyncContext(s.cxt.Copy())
}
// AsMap returns an unsynchronized map of the values in this context.
//
// This will give you access to the values, not the datasources or logger.
//
// The context is locked while the map is built and returned.
func (s *synchronizedContext) AsMap() map[string]ContextValue {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cxt.AsMap()
}
// Logger locks the context and then gets the logger.
//
// This will NOT prevent events from being written to the logger, but
// it will prevent other changes.
func (s *synchronizedContext) Logger(name string) (io.Writer, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cxt.Logger(name)
}

// AddLogger puts a new logger into the context's logging subsystem.
//
// The context is locked during insertion.
func (s *synchronizedContext) AddLogger(name string, logger io.Writer) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cxt.AddLogger(name, logger)
}
// RemoveLogger locks the context and then removes the logger from the context.
//
// This does NOT stop events from being logged to the logger, nor does it close
// the underlying writer. However, once the logger is removed, no further log
// messages will be written to it.
func (s *synchronizedContext) RemoveLogger(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cxt.RemoveLogger(name)
}

// Log sends a message to the underlying logger.
//
// This method is not synchronized. It is expected that the underlying
// logger will handle synchronization.
func (s *synchronizedContext) Log(prefix string, v ...interface{}) {
	s.cxt.Log(prefix, v...)
}
// Logf formats a message and sends it to the logger.
//
// This method is not synchronized, as it is assumed that the underlying logger
// will handle synchronization.
func (s *synchronizedContext) Logf(prefix, format string, v ...interface{}) {
	s.cxt.Logf(prefix, format, v...)
}



