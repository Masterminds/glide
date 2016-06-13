package altsrc

import (
	"fmt"
	"reflect"
	"time"

	"github.com/codegangsta/cli"
)

// MapInputSource implements InputSourceContext to return
// data from the map that is loaded.
type MapInputSource struct {
	valueMap map[string]interface{}
}

// Int returns an int from the map if it exists otherwise returns 0
func (fsm *MapInputSource) Int(name string) (int, error) {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(int)
		if !isType {
			return 0, incorrectTypeForFlagError(name, "int", otherGenericValue)
		}

		return otherValue, nil
	}

	return 0, nil
}

// Duration returns a duration from the map if it exists otherwise returns 0
func (fsm *MapInputSource) Duration(name string) (time.Duration, error) {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(time.Duration)
		if !isType {
			return 0, incorrectTypeForFlagError(name, "duration", otherGenericValue)
		}
		return otherValue, nil
	}

	return 0, nil
}

// Float64 returns an float64 from the map if it exists otherwise returns 0
func (fsm *MapInputSource) Float64(name string) (float64, error) {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(float64)
		if !isType {
			return 0, incorrectTypeForFlagError(name, "float64", otherGenericValue)
		}
		return otherValue, nil
	}

	return 0, nil
}

// String returns a string from the map if it exists otherwise returns an empty string
func (fsm *MapInputSource) String(name string) (string, error) {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(string)
		if !isType {
			return "", incorrectTypeForFlagError(name, "string", otherGenericValue)
		}
		return otherValue, nil
	}

	return "", nil
}

// StringSlice returns an []string from the map if it exists otherwise returns nil
func (fsm *MapInputSource) StringSlice(name string) ([]string, error) {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.([]string)
		if !isType {
			return nil, incorrectTypeForFlagError(name, "[]string", otherGenericValue)
		}
		return otherValue, nil
	}

	return nil, nil
}

// IntSlice returns an []int from the map if it exists otherwise returns nil
func (fsm *MapInputSource) IntSlice(name string) ([]int, error) {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.([]int)
		if !isType {
			return nil, incorrectTypeForFlagError(name, "[]int", otherGenericValue)
		}
		return otherValue, nil
	}

	return nil, nil
}

// Generic returns an cli.Generic from the map if it exists otherwise returns nil
func (fsm *MapInputSource) Generic(name string) (cli.Generic, error) {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(cli.Generic)
		if !isType {
			return nil, incorrectTypeForFlagError(name, "cli.Generic", otherGenericValue)
		}
		return otherValue, nil
	}

	return nil, nil
}

// Bool returns an bool from the map otherwise returns false
func (fsm *MapInputSource) Bool(name string) (bool, error) {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(bool)
		if !isType {
			return false, incorrectTypeForFlagError(name, "bool", otherGenericValue)
		}
		return otherValue, nil
	}

	return false, nil
}

// BoolT returns an bool from the map otherwise returns true
func (fsm *MapInputSource) BoolT(name string) (bool, error) {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(bool)
		if !isType {
			return true, incorrectTypeForFlagError(name, "bool", otherGenericValue)
		}
		return otherValue, nil
	}

	return true, nil
}

func incorrectTypeForFlagError(name, expectedTypeName string, value interface{}) error {
	valueType := reflect.TypeOf(value)
	valueTypeName := ""
	if valueType != nil {
		valueTypeName = valueType.Name()
	}

	return fmt.Errorf("Mismatched type for flag '%s'. Expected '%s' but actual is '%s'", name, expectedTypeName, valueTypeName)
}
