package jsonpointer

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/dustin/gojson"
)

var unparsable = errors.New("I can't parse this")

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

var decoder = strings.NewReplacer("~1", "/", "~0", "~")

func parsePointer(s string) []string {
	a := strings.Split(s[1:], "/")

	for i := range a {
		if strings.Contains(a[i], "~") {
			a[i] = decoder.Replace(a[i])
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

func grokLiteral(b []byte) string {
	s, ok := json.UnquoteBytes(b)
	if !ok {
		panic("could not grok literal " + string(b))
	}
	return string(s)
}

// Find a section of raw JSON by specifying a JSONPointer.
func Find(data []byte, path string) ([]byte, error) {
	if path == "" {
		return data, nil
	}

	needle := parsePointer(path)

	scan := &json.Scanner{}
	scan.Reset()

	offset := 0
	beganLiteral := 0
	current := []string{}
	for {
		if offset >= len(data) {
			break
		}
		newOp := scan.Step(scan, int(data[offset]))
		offset++

		switch newOp {
		case json.ScanBeginArray:
			current = append(current, "0")
		case json.ScanObjectKey:
			current = append(current, grokLiteral(data[beganLiteral-1:offset-1]))
		case json.ScanBeginLiteral:
			beganLiteral = offset
		case json.ScanArrayValue:
			n, err := strconv.Atoi(current[len(current)-1])
			if err != nil {
				return nil, err
			}
			current[len(current)-1] = strconv.Itoa(n + 1)
		case json.ScanObjectValue:
			current = current[:len(current)-1]
		case json.ScanEndArray:
			current = current[:len(current)-1]
		}

		if (newOp == json.ScanBeginArray || newOp == json.ScanArrayValue ||
			newOp == json.ScanObjectKey) && arreq(needle, current) {
			val, _, err := json.NextValue(data[offset:], scan)
			return val, err
		}
	}

	return nil, nil
}

// List all possible pointers from the given input.
func ListPointers(data []byte) ([]string, error) {
	rv := []string{""}

	scan := &json.Scanner{}
	scan.Reset()

	offset := 0
	beganLiteral := 0
	var current []string
	for {
		if offset >= len(data) {
			break
		}
		newOp := scan.Step(scan, int(data[offset]))
		offset++

		switch newOp {
		case json.ScanBeginArray:
			current = append(current, "0")
		case json.ScanObjectKey:
			current = append(current, grokLiteral(data[beganLiteral-1:offset-1]))
		case json.ScanBeginLiteral:
			beganLiteral = offset
		case json.ScanArrayValue:
			n, err := strconv.Atoi(current[len(current)-1])
			if err != nil {
				return nil, err
			}
			current[len(current)-1] = strconv.Itoa(n + 1)
		case json.ScanObjectValue, json.ScanEndArray, json.ScanEndObject:
			current = current[:len(current)-1]
		}

		if newOp == json.ScanBeginArray || newOp == json.ScanArrayValue ||
			newOp == json.ScanObjectKey {
			rv = append(rv, encodePointer(current))
		}
	}

	return rv, nil
}

// Find a section of raw JSON by specifying a JSONPointer.
func FindMany(data []byte, paths []string) (map[string][]byte, error) {
	tpaths := make([]string, 0, len(paths))
	m := map[string][]byte{}
	for _, p := range paths {
		if p == "" {
			m[p] = data
		} else {
			tpaths = append(tpaths, p)
		}
	}
	sort.Strings(tpaths)

	scan := &json.Scanner{}
	scan.Reset()

	offset := 0
	todo := len(tpaths)
	beganLiteral := 0
	matchedAt := 0
	var current []string
	for todo > 0 {
		if offset >= len(data) {
			break
		}
		newOp := scan.Step(scan, int(data[offset]))
		offset++

		switch newOp {
		case json.ScanBeginArray:
			current = append(current, "0")
		case json.ScanObjectKey:
			current = append(current, grokLiteral(data[beganLiteral-1:offset-1]))
		case json.ScanBeginLiteral:
			beganLiteral = offset
		case json.ScanArrayValue:
			n, err := strconv.Atoi(current[len(current)-1])
			if err != nil {
				return nil, err
			}
			current[len(current)-1] = strconv.Itoa(n + 1)
		case json.ScanObjectValue, json.ScanEndArray, json.ScanEndObject:
			current = current[:len(current)-1]
		}

		if newOp == json.ScanBeginArray || newOp == json.ScanArrayValue ||
			newOp == json.ScanObjectKey {

			if matchedAt < len(current)-1 {
				continue
			}
			if matchedAt > len(current) {
				matchedAt = len(current)
			}

			currentStr := encodePointer(current)
			off := sort.SearchStrings(tpaths, currentStr)
			if off < len(tpaths) {
				// Check to see if the path we're
				// going down could even lead to a
				// possible match.
				if strings.HasPrefix(tpaths[off], currentStr) {
					matchedAt++
				}
				// And if it's not an exact match, keep parsing.
				if tpaths[off] != currentStr {
					continue
				}
			} else {
				// Fell of the end of the list, no possible match
				continue
			}

			// At this point, we have an exact match, so grab it.
			stmp := &json.Scanner{}
			val, _, err := json.NextValue(data[offset:], stmp)
			if err != nil {
				return m, err
			}
			m[currentStr] = val
			todo--
		}
	}

	return m, nil
}
