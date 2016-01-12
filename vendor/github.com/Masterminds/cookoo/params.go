package cookoo

type Params struct {
	storage map[string]interface{}
}

// NewParamsWithValues initializes a Params object with the given values.
//
// Create a new Params instance, initialized with the given map.
// Note that the given map is actually used (not copied).
func NewParamsWithValues(initialVals map[string]interface{}) *Params {
	p := new(Params)
	p.Init(initialVals)
	return p
}

// NewParams creates a Params object of a given size.
func NewParams(size int) *Params {
	p := new(Params)
	p.storage = make(map[string]interface{}, size)
	return p
}

// Init initializes a Params object with an initial map of values.
func (p *Params) Init(initialValues map[string]interface{}) {
	p.storage = initialValues
}

func (p *Params) set(name string, value interface{}) bool {
	_, ret := p.storage[name]
	p.storage[name] = value
	return ret
}

// Has checks if a parameter exists, and return it if found.
func (p *Params) Has(name string) (value interface{}, ok bool) {
	value, ok = p.storage[name]
	if value == nil {
		ok = false
	}
	return
}

// Get gets a parameter value, or returns the default value.
func (p *Params) Get(name string, defaultValue interface{}) interface{} {
	val, ok := p.Has(name)
	if ok {
		return val
	}
	return defaultValue
}

// Requires verifies that the given keys exist in the Params.
//
// Require that a given list of parameters are present.
// If they are all present, ok = true. Otherwise, ok = false and the
// `missing` array contains a list of missing params.
func (p *Params) Requires(paramNames ...string) (ok bool, missing []string) {
	missing = make([]string, 0, len(p.storage))
	for _, val := range paramNames {
		_, ok := p.storage[val]
		if !ok {
			missing = append(missing, val)
		}

	}
	ok = len(missing) == 0
	return
}

// RequiresValue verifies that the given keys exist and that their values are
// non-empty.
//
// Requires that given parameters are present and non-empty.
// This is more powerful than Requires(), which simply checks to see if the the Using() clause declared
// the value.
func (p *Params) RequiresValue(paramNames ...string) (ok bool, missing []string) {
	missing = make([]string, 0, len(p.storage))
	for _, val := range paramNames {
		vv, ok := p.storage[val]
		switch vv.(type) {
		default:
			if vv == nil {
				ok = false
			}
		case string:
			if vv == nil || len(vv.(string)) == 0 {
				ok = false
			}
		case []interface{}:
			if vv == nil || len(vv.([]interface{})) == 0 {
				ok = false
			}
		case map[interface{}]interface{}:
			if vv == nil || len(vv.(map[interface{}]interface{})) == 0 {
				ok = false
			}
		}

		if !ok {
			missing = append(missing, val)
		}
	}
	ok = len(missing) == 0
	return
}

// AsMap returns all parameters as a map[string]interface{}.
//
// This does no checking of the parameters.
func (p *Params) AsMap() map[string]interface{} {
	return p.storage
}

// Len returns the number of params.
func (p *Params) Len() int {
	return len(p.storage)
}

// Validate provides a validator callback for params.
// Given a name and a validation function, return a valid value.
// If the value is not valid, ok = false.
func (p *Params) Validate(name string, validator func(interface{}) bool) (value interface{}, ok bool) {
	value, ok = p.storage[name]
	if !ok {
		return
	}

	if !validator(value.(interface{})) {
		// XXX: For safety, we set a failed value to nil.
		value = nil
		ok = false
	}
	return
}
