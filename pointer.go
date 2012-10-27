package jsonpointer

import (
	"errors"
	"strconv"
	"strings"
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
	a := strings.Split(s, "/")

	for i := range a {
		if strings.Contains(a[i], "~") {
			a[i] = strings.Replace(a[i], "~1", "/", -1)
			a[i] = strings.Replace(a[i], "~0", "~", -1)
		}
	}
	return a
}

func grokLiteral(b []byte) string {
	s, ok := unquoteBytes(b)
	if !ok {
		panic("could not grok literal " + string(b))
	}
	return string(s)
}

// Find a section of raw JSON by specifying a JSONPointer.
func Find(data []byte, path string) ([]byte, error) {
	if path == "" {
		return RawMessage(data), nil
	}

	needle := parsePointer(path[1:])

	scanner := &scanner{}
	scanner.reset()

	offset := 0
	beganLiteral := 0
	current := []string{}
	for {
		var newOp int
		if offset >= len(data) {
			newOp = scanner.eof()
			break
			offset = len(data) + 1 // mark processed EOF with len+1
		} else {
			c := int(data[offset])
			offset++
			newOp = scanner.step(scanner, c)
		}

		switch newOp {
		case scanBeginArray:
			current = append(current, "0")
		case scanObjectKey:
			current = append(current, grokLiteral(data[beganLiteral-1:offset-1]))
		case scanBeginLiteral:
			beganLiteral = offset
		case scanArrayValue:
			n, err := strconv.Atoi(current[len(current)-1])
			if err != nil {
				return nil, err
			}
			current[len(current)-1] = strconv.Itoa(n + 1)
		case scanObjectValue:
			current = current[:len(current)-1]
		case scanEndArray:
			current = current[:len(current)-1]
		}

		if arreq(needle, current) {
			val, _, err := nextValue(data[offset:], scanner)
			return val, err
		}
	}

	return nil, unparsable
}

// Find a section of raw JSON by specifying a JSONPointer.
func FindMany(data []byte, paths []string) (map[string][]byte, error) {
	rv := map[string][]byte{}
	for _, p := range paths {
		d, err := Find(data, p)
		if err != nil {
			return rv, err
		}
		rv[p] = d
	}
	return rv, nil
}
