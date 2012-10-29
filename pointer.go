package jsonpointer

import (
	"errors"
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
	todo := map[string]bool{}
	m := map[string][]byte{}
	for _, p := range paths {
		if p == "" {
			m[p] = data
		} else {
			todo[p] = true
		}
	}

	scan := &json.Scanner{}
	scan.Reset()

	offset := 0
	beganLiteral := 0
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
			currentStr := encodePointer(current)
			if !todo[currentStr] {
				continue
			}

			stmp := &json.Scanner{}
			val, _, err := json.NextValue(data[offset:], stmp)
			if err != nil {
				return m, err
			}
			m[currentStr] = val
			delete(todo, currentStr)
			if len(todo) == 0 {
				return m, nil
			}
		}
	}

	return m, nil
}
