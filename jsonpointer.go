// JSON Pointer
//
// Path access to map[string]interface{} objects (typical of JSON
// decoding).
package jsonpointer

import (
	"strconv"
	"strings"
)

func unescape(s string) string {
	if strings.Contains(s, "~") {
		s = strings.Replace(s, "~1", "/", -1)
		s = strings.Replace(s, "~0", "~", -1)
	}
	return s
}

// Get the value at the specified path.
func Get(m map[string]interface{}, path string) interface{} {
	if path == "" {
		return m
	}

	parts := strings.Split(path[1:], "/")
	var rv interface{} = m

	for _, p := range parts {
		p = unescape(p)
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
