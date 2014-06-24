// Package jsonpointer implements RFC6901 JSON Pointers
package jsonpointer

import (
	"errors"
	"strconv"
	"strings"
)

// methods are not thread safe.
// Get(), Set(), Delete() are idempotent ops.

// ErrorInvalidPath specified path does not exist within the document.
var ErrorInvalidPath = errors.New("jsonpointer.invalidPath")

// ErrorInvalidType specified path does not refer to expected type.
var ErrorInvalidType = errors.New("jsonpointer.invalidType")

type mHandler func(old interface{}) (latest interface{}, err error)

// Get the value at the specified path, if returned value is `nil` then
// specified path in invalid.
func Get(m map[string]interface{}, path string) (rv interface{}) {
	switch path {
	case "":
		return m

	default:
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

// Incr increments referenced field's value by `val`.
// not an idempotent operation.
func Incr(m map[string]interface{}, path string, val int) error {
	return mutateField(m, path, func(old interface{}) (interface{}, error) {
		if value, ok := old.(float64); ok {
			return value + float64(val), nil
		}
		return nil, ErrorInvalidType
	})
}

// Incrs increments referenced field's array elements by list of vals.
// not an idempotent operation.
func Incrs(m map[string]interface{}, path string, vals ...int) error {
	return mutateField(m, path, func(old interface{}) (interface{}, error) {
		if values, ok := old.([]interface{}); ok {
			for i, val := range vals {
				values[i] = values[i].(float64) + float64(val)
			}
			return values, nil
		}
		return nil, ErrorInvalidType
	})
}

// Decr decrements referenced field's value by `val`.
// not an idempotent operation.
func Decr(m map[string]interface{}, path string, val int) error {
	return mutateField(m, path, func(old interface{}) (interface{}, error) {
		if value, ok := old.(float64); ok {
			return value - float64(val), nil
		}
		return nil, ErrorInvalidType
	})
}

func mutateField(m map[string]interface{}, path string, fn mHandler) error {
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
		latestV, err := fn(v[last])
		if err != nil {
			return err
		}
		v[last] = latestV

	case []interface{}:
		if i, err := strconv.Atoi(last); err == nil && i < len(v) {
			latestV, err := fn(v[i])
			if err != nil {
				return err
			}
			v[i] = latestV
		} else {
			return ErrorInvalidPath
		}

	default:
		return ErrorInvalidPath
	}
	return nil
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
