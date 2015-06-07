package jsonpointer

import (
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

func handleTilde(p string) string {
	if strings.Contains(p, "~") {
		p = strings.Replace(p, "~1", "/", -1)
		p = strings.Replace(p, "~0", "~", -1)
	}
	return p
}

// Get the value at the specified path.
func Get(m map[string]interface{}, path string) interface{} {
	if path == "" {
		return m
	}

	parts := strings.Split(path[1:], "/")
	var rv interface{} = m

	for _, p := range parts {

		switch v := rv.(type) {
		case map[string]interface{}:
			handleTilde(p)
			rv = v[p]
		case bson.M:
			handleTilde(p)
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
