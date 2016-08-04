package jsonpointer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

func arreq(a, b []string) bool {
	if len(a) == len(b) {
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}

	return false
}

// unescape unescapes a tilde escaped string.
//
// It's dumb looking, but it benches faster than strings.NewReplacer
func unescape(s string) string {
	return strings.Replace(strings.Replace(s, "~0", "~", -1), "~1", "/", -1)
}

func parsePointer(s string) []string {
	a := strings.Split(s[1:], "/")
	if !strings.Contains(s, "~") {
		return a
	}

	for i := range a {
		if strings.Contains(a[i], "~") {
			a[i] = unescape(a[i])
		}
	}
	return a
}

func encodePointer(p []string) string {
	out := make([]rune, 0, 64)

	for _, s := range p {
		out = append(out, '/')
		for _, c := range s {
			switch c {
			case '/':
				out = append(out, '~', '1')
			case '~':
				out = append(out, '~', '0')
			default:
				out = append(out, c)
			}
		}
	}
	return string(out)
}

// FindDecode finds an object by JSONPointer path and then decode the
// result into a user-specified object.  Errors if a properly
// formatted JSON document can't be found at the given path.
func FindDecode(data []byte, path string, into interface{}) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()

	if path == "" {
		return d.Decode(into)
	}

	needle := parsePointer(path)

	var derr error
	err := visit(data, func(current []string, d *json.Decoder) bool {
		if arreq(current, needle) {
			derr = d.Decode(into)
			return false
		}
		return true
	})

	if derr != nil {
		err = derr
	}

	return err
}

const (
	array  = "array"
	object = "object"
)

// Find a section of raw JSON by specifying a JSONPointer.
func Find(data []byte, path string) ([]byte, error) {
	var m json.RawMessage
	err := FindDecode(data, path, &m)
	if err == io.EOF {
		return nil, nil
	}
	return []byte(m), err
}

func sliceToEnd(s []string) []string {
	end := len(s) - 1
	if end >= 0 {
		s = s[:end]
	}
	return s

}

func mustParseInt(s string) int {
	n, err := strconv.Atoi(s)
	if err == nil {
		return n
	}
	panic(err)
}

func visit(data []byte, f func(current []string, d *json.Decoder) bool) error {
	if len(data) == 0 {
		return fmt.Errorf("Invalid JSON")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()

	var current []string
	var types []string
	wantKey := false
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}
		typ := ""
		if len(types) > 0 {
			typ = types[len(types)-1]
		}

		switch v := t.(type) {
		case json.Delim:
			switch v {
			case '[':
				current = append(current, "0")
				types = append(types, array)
				if d.More() && !f(current, d) {
					return nil
				}
				continue
			case '{':
				wantKey = true
				types = append(types, object)
				current = append(current, "")
				continue
			case '}', ']':
				current = sliceToEnd(current)
				types = sliceToEnd(types)
				ptyp := ""
				if len(types) > 0 {
					ptyp = types[len(types)-1]
				}
				switch ptyp {
				case object:
					wantKey = true
				case array:
					n := mustParseInt(current[len(current)-1])
					current[len(current)-1] = strconv.Itoa(n + 1)
					if d.More() && !f(current, d) {
						return nil
					}
				}
				continue
			}
		case json.Number:
		case bool:
		case string:
			if wantKey {
				current[len(current)-1] = v
				if !f(current, d) {
					return nil
				}
			}
		case nil:
		}

		switch typ {
		case object:
			wantKey = !wantKey
		case array:
			n := mustParseInt(current[len(current)-1])
			current[len(current)-1] = strconv.Itoa(n + 1)
			if d.More() && !f(current, d) {
				return nil
			}
		}

	}
}

func ListPointers(data []byte) ([]string, error) {
	rv := []string{""}
	err := visit(data, func(current []string, d *json.Decoder) bool {
		rv = append(rv, encodePointer(current))
		return true
	})
	if err == io.EOF {
		err = nil
	}
	return rv, err
}

// FindMany finds several jsonpointers in one pass through the input.
func FindMany(data []byte, paths []string) (map[string][]byte, error) {
	tpaths := make([]string, 0, len(paths))
	m := map[string][]byte{}
	for _, p := range paths {
		if p == "" {
			m[p] = data
		} else {
			rv, err := Find(data, p)
			if err != nil {
				return m, err
			}
			if len(rv) > 0 {
				m[p] = rv
			}
		}
	}
	sort.Strings(tpaths)

	return m, nil
}
