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

func parsePointer(s string) []string {
	a := strings.Split(s[1:], "/")

	for i := range a {
		if strings.Contains(a[i], "~") {
			a[i] = strings.Replace(a[i], "~1", "/", -1)
			a[i] = strings.Replace(a[i], "~0", "~", -1)
		}
	}
	return a
}

func encodePointer(p []string) string {
	a := make([]string, 0, len(p))
	for _, s := range p {
		s = strings.Replace(s, "~", "~0", -1)
		s = strings.Replace(s, "/", "~1", -1)
		a = append(a, s)
	}
	return "/" + strings.Join(a, "/")
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

	scanner := &json.Scanner{}
	scanner.Reset()

	offset := 0
	beganLiteral := 0
	current := []string{}
	for {
		var newOp int
		if offset >= len(data) {
			newOp = scanner.EOF()
			break
			offset = len(data) + 1 // mark processed EOF with len+1
		} else {
			c := int(data[offset])
			offset++
			newOp = scanner.Step(scanner, c)
		}

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
			val, _, err := json.NextValue(data[offset:], scanner)
			return val, err
		}
	}

	return nil, nil
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
	for {
		var newOp int
		if offset >= len(data) {
			newOp = scan.EOF()
			break
			offset = len(data) + 1 // mark processed EOF with len+1
		} else {
			c := int(data[offset])
			offset++
			newOp = scan.Step(scan, c)
		}

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
			if todo == 0 {
				return m, nil
			}
		}
	}

	return m, nil
}
