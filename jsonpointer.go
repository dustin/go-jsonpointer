package jsonpointer

import (
	"net/url"
	"strconv"
	"strings"
)

// Get the value at the specified path.
func Get(m map[string]interface{}, path string) interface{} {
	if path == "/" {
		return m
	}

	parts := strings.Split(path, "/")
	var rv interface{} = m
	var err error

	for _, p := range parts {
		if p == "" {
			continue
		}
		p, err = url.QueryUnescape(p)
		if err != nil {
			return nil
		}
		switch v := rv.(type) {
		case map[string]interface{}:
			rv = v[p]
		case []interface{}:
			i, err := strconv.Atoi(p)
			if err == nil && i < len(v) {
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
