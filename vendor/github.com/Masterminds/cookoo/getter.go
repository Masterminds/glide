package cookoo

/*
This file contains mainly convenience functions for working with Params,
Context, and KeyValueDatasource.

Many of these functions are forward-looking to Cookoo 2.x, where Context and
KeyValueDatasource will both be Getter implementations.
*/

import (
	"reflect"
)

// Getter can get values in two ways.
//
// A Get() can be given a default value, in which case it will return either
// the value associated with the key or, if that's not found, the default value.
//
// A Has() doesn't take a default value, but instead returns both the value
// (if found) and a boolean flag indicating whether it is found.
//
// In Cookoo 1.x, Context's Get() function returns a ContextValue instead of
// an interface. For that reason, you may need to wrap Cxt in GettableCxt
// to make it a true Getter.
//
// In Cookoo 1.x, KeyValueDatasource uses Value() instead of Get()/Has(). For
// that reason, you can wrap a KeyValueDatasource in a GettableDS() to make
// it behave like a Getter.
type Getter interface {
	Get(string, interface{}) interface{}
	Has(string) (interface{}, bool)
}

// GettableDS makes a KeyValueDatasource into a Getter.
//
// This is forward-compatibility code, and will be rendered unnecessary in
// Cookoo 2.x.
func GettableDS(ds KeyValueDatasource) Getter {
	return &gettableDatasource{ds}
}

// GettableCxt makes a Context into a Getter.
//
// This is forward-compatibility code, and will be rendered unnecessary in
// Cookoo 2.x.
func GettableCxt(cxt Context) Getter {
	return &gettableContext{cxt}
}

// GettableDatasource Makes a KeyValueDatasource match the Getter interface.
//
// In future versions of Cookoo, core Datasources will directly implement Getter.
type gettableDatasource struct {
	KeyValueDatasource
}

func (g *gettableDatasource) Get(key string, defaultVal interface{}) interface{} {
	ret := g.KeyValueDatasource.Value(key)
	if ret == nil || !reflect.ValueOf(ret).IsValid() {
		return defaultVal
	}
	return ret
}

func (g *gettableDatasource) Has(key string) (interface{}, bool) {
	ret := g.KeyValueDatasource.Value(key)
	if ret == nil || !reflect.ValueOf(ret).IsValid() {
		return nil, false
	}
	return ret, true
}

// GettableContext wraps a context and makes it a Getter.
// Since Context returns ContextValue objects, we have to write this stupid wrapper.
type gettableContext struct {
	Context
}

func (g *gettableContext) Get(key string, defaultVal interface{}) interface{} {
	return g.Context.Get(key, defaultVal)
}

func (g *gettableContext) Has(key string) (interface{}, bool) {
	return g.Context.Has(key)
}

// GetString is a convenience function for getting strings.
//
// This simplifies getting strings from a Context, a Params, or a
// GettableDatasource.
func GetString(key, defaultValue string, source Getter) string {
	out := source.Get(key, defaultValue)
	ret, ok := out.(string)
	if !ok {
		return defaultValue
	}
	return ret
}

// GetBool gets a boolean value from any Getter.
func GetBool(key string, defaultValue bool, source Getter) bool {
	out := source.Get(key, defaultValue)
	ret, ok := out.(bool)
	if !ok {
		return defaultValue
	}
	return ret
}

// GetInt gets an int from any Getter.
func GetInt(key string, defaultValue int, source Getter) int {
	out := source.Get(key, defaultValue)
	ret, ok := out.(int)
	if !ok {
		return defaultValue
	}
	return ret
}

// GetInt64 gets an int64 from any Getter.
func GetInt64(key string, defaultValue int64, source Getter) int64 {
	out := source.Get(key, defaultValue)
	ret, ok := out.(int64)
	if !ok {
		return defaultValue
	}
	return ret
}

// GetInt32 gets an int32 from any Getter.
func GetInt32(key string, defaultValue int32, source Getter) int32 {
	out := source.Get(key, defaultValue)
	ret, ok := out.(int32)
	if !ok {
		return defaultValue
	}
	return ret
}

// GetUint64 gets a uint64 from any Getter.
func GetUint64(key string, defaultVal uint64, source Getter) uint64 {
	out := source.Get(key, defaultVal)
	ret, ok := out.(uint64)
	if !ok {
		return defaultVal
	}
	return ret
}

// GetFloat64 gets a float64 from any Getter.
func GetFloat64(key string, defaultVal float64, source Getter) float64 {
	out := source.Get(key, defaultVal)
	ret, ok := out.(float64)
	if !ok {
		return defaultVal
	}
	return ret
}

// HasString is a convenience function to perform Has() and return a string.
func HasString(key string, source Getter) (string, bool) {
	v, ok := source.Has(key)
	if !ok {
		return "", ok
	}
	strval, kk := v.(string)
	if !kk {
		return "", kk
	}
	return strval, kk
}

// HasBool returns the value and a flag indicated whether the flag value was found.
//
// Default value is false if ok is false.
func HasBool(key string, source Getter) (bool, bool) {
	v, ok := source.Has(key)
	if !ok {
		return false, ok
	}
	strval, kk := v.(bool)
	if !kk {
		return false, kk
	}
	return strval, kk
}

// HasInt returns the int value for key, and a flag indicated if it was found.
//
// If ok is false, the int value will be 0
func HasInt(key string, source Getter) (int, bool) {
	v, ok := source.Has(key)
	if !ok {
		return 0, ok
	}
	val, kk := v.(int)
	if !kk {
		return 0, kk
	}
	return val, kk
}

// HasInt64 returns the int64 value for key, and a flag indicated if it was found.
//
// If ok is false, the int value will be 0
func HasInt64(key string, source Getter) (int64, bool) {
	v, ok := source.Has(key)
	if !ok {
		return 0, ok
	}
	val, kk := v.(int64)
	if !kk {
		return 0, kk
	}
	return val, kk
}

// HasInt32 returns the int32 value for key, and a flag indicated if it was found.
//
// If ok is false, the int value will be 0
func HasInt32(key string, source Getter) (int32, bool) {
	v, ok := source.Has(key)
	if !ok {
		return 0, ok
	}
	val, kk := v.(int32)
	if !kk {
		return 0, kk
	}
	return val, kk
}

// HasUint64 returns the uint64 value for key, and a flag indicated if it was found.
//
// If ok is false, the int value will be 0
func HasUint64(key string, source Getter) (uint64, bool) {
	v, ok := source.Has(key)
	if !ok {
		return 0, ok
	}
	val, kk := v.(uint64)
	if !kk {
		return 0, kk
	}
	return val, kk
}

// HasFloat64 returns the float64 value for key, and a flag indicated if it was found.
//
// If ok is false, the float value will be 0
func HasFloat64(key string, source Getter) (float64, bool) {
	v, ok := source.Has(key)
	if !ok {
		return 0, ok
	}
	val, kk := v.(float64)
	if !kk {
		return 0, kk
	}
	return val, kk
}

// GetFromFirst gets the value from the first Getter that has the key.
//
// This provides a method for scanning, for example, Params, Context, and
// KeyValueDatasource and returning the first one that matches.
//
// If no Getter has the key, the default value is returned, and the returned
// Getter is an instance of DefaultGetter.
func GetFromFirst(key string, defaultVal interface{}, sources ...Getter) (interface{}, Getter) {
	for _, s := range sources {
		val, ok := s.Has(key)
		if ok {
			return val, s
		}
	}

	return defaultVal, &DefaultGetter{defaultVal}
}

// DefaultGetter represents a Getter instance for a default value.
//
// A default getter always returns the given default value.
type DefaultGetter struct {
	val interface{}
}

func (e *DefaultGetter) Get(name string, value interface{}) interface{} {
	return e.val
}
func (e *DefaultGetter) Has(name string) (interface{}, bool) {
	return e.val, true
}
