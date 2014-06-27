// Package jsonpointer implements RFC6901 JSON Pointers.
package jsonpointer

import (
	"errors"
	"strconv"
	"strings"
)

// methods are not thread safe.
// Get(), Set(), Delete() are idempotent ops.

// ErrorInvalidPath is returned when specified path does not refer to a valid
// field within the document.
var ErrorInvalidPath = errors.New("jsonpointer.invalidPath")

// ErrorInvalidType is returned when specified path does not refer to expected
// type.
var ErrorInvalidType = errors.New("jsonpointer.invalidType")

type mHandler func(old interface{}) (latest interface{}, err error)

// Get the value at the specified path, if returned value is `nil` then
// specified path in invalid.
func Get(m map[string]interface{}, path string) (rv interface{}) {
	if path == "" {
		return m
	}

	parts := strings.Split(path[1:], "/")
	rv = m
	for _, p := range parts {
		switch v := rv.(type) {
		case map[string]interface{}:
			if strings.Contains(p, "~") {
				p = strings.Replace(p, "~1", "/", -1)
				p = strings.Replace(p, "~0", "~", -1)
			}
			rv = v[p]

		case []interface{}:
			if i, err := strconv.Atoi(p); err == nil && i < len(v) {
				rv = v[i]
			} else {
				return nil
			}

		default:
			return nil
		}
	}
	return rv
}

// Set value at the specified path. `""` is a special path that needs to be
// handled by caller them self.
func Set(m map[string]interface{}, path string, value interface{}) (err error) {
	container, last, err := getContainer(m, path)
	if err != nil {
		return err
	}

	if strings.Contains(last, "~") {
		last = strings.Replace(last, "~1", "/", -1)
		last = strings.Replace(last, "~0", "~", -1)
	}

	switch v := container.(type) {
	case map[string]interface{}:
		v[last] = value // idempotent operation

	case []interface{}:
		if i, err := strconv.Atoi(last); err == nil && i < len(v) {
			v[i] = value // idempotent operation
		} else {
			return ErrorInvalidPath
		}

	default:
		return ErrorInvalidPath
	}
	return nil
}

// Delete the value at specified path. To keep this an idempotent
// operation, path cannot end indexing into array type. `""` is a special path
// that needs to be handled by caller them self.
func Delete(m map[string]interface{}, path string) (err error) {
	container, last, err := getContainer(m, path)
	if err != nil {
		return err
	}

	if strings.Contains(last, "~") {
		last = strings.Replace(last, "~1", "/", -1)
		last = strings.Replace(last, "~0", "~", -1)
	}

	switch v := container.(type) {
	case map[string]interface{}:
		delete(v, last) // idempotent operation

	default:
		return ErrorInvalidPath
	}
	return nil
}

// DeleteAny delete the value at specified path, path can end indexing
// into array type, hence not an idempotent operation. `""` is a special path
// that needs to be handled by caller them self.
func DeleteAny(m map[string]interface{}, path string) (err error) {
	container, last, err := getContainer(m, path)
	if err != nil {
		return err
	}

	if strings.Contains(last, "~") {
		last = strings.Replace(last, "~1", "/", -1)
		last = strings.Replace(last, "~0", "~", -1)
	}

	switch v := container.(type) {
	case map[string]interface{}:
		delete(v, last) // idempotent operation

	case []interface{}:
		if i, err := strconv.Atoi(last); (err == nil) && (i < len(v)) {
			copy(v[i:], v[i+1:])
			v[len(v)-1] = nil
			v = v[:len(v)-1]
			parts := parsePointer(path)
			Set(m, "/"+strings.Join(parts[:len(parts)-1], "/"), v)
		} else {
			return ErrorInvalidPath
		}

	default:
		return ErrorInvalidPath
	}
	return nil
}

// Incr increments referenced field's value(s) by list of `vals`. To increment
// a value of type float64 len(vals) must be 1, otherwise referred field
// is expected to be []float64. This is not an idempotent operation.
func Incr(m map[string]interface{}, path string, vals ...int) error {
	return mutateField(m, path, func(old interface{}) (interface{}, error) {
		switch value := old.(type) {
		case float64:
			if len(vals) == 1 {
				return value + float64(vals[0]), nil
			}
			return nil, ErrorInvalidType

		case []interface{}:
			for i, val := range vals {
				value[i] = value[i].(float64) + float64(val)
			}
			return value, nil

		case []float64:
			for i, val := range vals {
				value[i] = value[i] + float64(val)
			}
			return value, nil
		}
		return nil, ErrorInvalidType
	})
}

// Decr decrements referenced field's value(s) by list of `vals`. To decrement
// a value of type float64 len(vals) must be 1, otherwise referred field
// is expected to be []float64. This is not an idempotent operation.
func Decr(m map[string]interface{}, path string, vals ...int) error {
	return mutateField(m, path, func(old interface{}) (interface{}, error) {
		switch value := old.(type) {
		case float64:
			if len(vals) == 1 {
				return value - float64(vals[0]), nil
			}
			return nil, ErrorInvalidType

		case []interface{}:
			for i, val := range vals {
				value[i] = value[i].(float64) - float64(val)
			}
			return value, nil

		case []float64:
			for i, val := range vals {
				value[i] = value[i] - float64(val)
			}
			return value, nil
		}
		return nil, ErrorInvalidType
	})
}

func mutateField(m map[string]interface{}, path string, fn mHandler) error {
	var err error
	var latestV interface{}

	container, last, err := getContainer(m, path)
	if err != nil {
		return err
	}

	if strings.Contains(last, "~") {
		last = strings.Replace(last, "~1", "/", -1)
		last = strings.Replace(last, "~0", "~", -1)
	}

	switch v := container.(type) {
	case map[string]interface{}:
		if latestV, err = fn(v[last]); err == nil {
			v[last] = latestV
		}

	case []float64:
		var i int
		if i, err = strconv.Atoi(last); err == nil && i < len(v) {
			if latestV, err = fn(v[i]); err == nil {
				v[i] = latestV.(float64)
			}
		} else {
			err = ErrorInvalidPath
		}

	case []interface{}:
		var i int
		if i, err = strconv.Atoi(last); err == nil && i < len(v) {
			if latestV, err = fn(v[i]); err == nil {
				v[i] = latestV
			}
		} else {
			err = ErrorInvalidPath
		}

	default:
		err = ErrorInvalidPath
	}
	return err
}

func getContainer(m map[string]interface{}, path string) (container interface{}, last string, err error) {
	parts := strings.Split(path[1:], "/")
	l := len(parts)

	container = m

	hs, last := parts[:l-1], parts[l-1]
	for _, h := range hs {
		switch v := container.(type) {
		case map[string]interface{}:
			if strings.Contains(h, "~") {
				h = strings.Replace(h, "~1", "/", -1)
				h = strings.Replace(h, "~0", "~", -1)
			}
			container = v[h]

		case []interface{}:
			if i, err := strconv.Atoi(h); err == nil && i < len(v) {
				container = v[i]
			} else {
				return container, last, ErrorInvalidPath
			}

		default:
			return container, last, ErrorInvalidPath
		}
	}
	return container, last, nil
}
