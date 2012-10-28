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

	needle := parsePointer(path)

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

		if (newOp == scanBeginArray || newOp == scanArrayValue ||
			newOp == scanObjectKey) && arreq(needle, current) {
			val, _, err := nextValue(data[offset:], scanner)
			return val, err
		}
	}

	return nil, unparsable
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

	scan := &scanner{}
	scan.reset()

	offset := 0
	beganLiteral := 0
	var current []string
	currentStr := ""
	for {
		var newOp int
		if offset >= len(data) {
			newOp = scan.eof()
			break
			offset = len(data) + 1 // mark processed EOF with len+1
		} else {
			c := int(data[offset])
			offset++
			newOp = scan.step(scan, c)
		}

		switch newOp {
		case scanBeginArray:
			current = append(current, "0")
			currentStr = encodePointer(current)
		case scanObjectKey:
			current = append(current, grokLiteral(data[beganLiteral-1:offset-1]))
			currentStr = encodePointer(current)
		case scanBeginLiteral:
			beganLiteral = offset
		case scanArrayValue:
			n, err := strconv.Atoi(current[len(current)-1])
			if err != nil {
				return nil, err
			}
			current[len(current)-1] = strconv.Itoa(n + 1)
			currentStr = encodePointer(current)
		case scanObjectValue, scanEndArray:
			current = current[:len(current)-1]
			currentStr = encodePointer(current)
		}

		if (newOp == scanBeginArray || newOp == scanArrayValue ||
			newOp == scanObjectKey) && todo[currentStr] {

			stmp := &scanner{}
			val, _, err := nextValue(data[offset:], stmp)
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
