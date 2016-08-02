package jsonpointer

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/dustin/gojson"
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

func grokLiteral(b []byte) string {
	s, ok := unquoteBytes(b)
	if !ok {
		panic("could not grok literal " + string(b))
	}
	return string(s)
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

// FindDecode finds an object by JSONPointer path and then decode the
// result into a user-specified object.  Errors if a properly
// formatted JSON document can't be found at the given path.
func FindDecode(data []byte, path string, into interface{}) error {
	d, err := Find(data, path)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, into)
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
			current[len(current)-1] = grokLiteral(data[beganLiteral-1 : offset-1])
		case json.ScanBeginLiteral:
			beganLiteral = offset
		case json.ScanArrayValue:
			n := mustParseInt(current[len(current)-1])
			current[len(current)-1] = strconv.Itoa(n + 1)
		case json.ScanEndArray, json.ScanEndObject:
			current = sliceToEnd(current)
		case json.ScanBeginObject:
			current = append(current, "")
		case json.ScanContinue, json.ScanSkipSpace, json.ScanObjectValue, json.ScanEnd:
		default:
			return nil, fmt.Errorf("found unhandled json op: %v", newOp)
		}

		if (newOp == json.ScanBeginArray || newOp == json.ScanArrayValue ||
			newOp == json.ScanObjectKey) && arreq(needle, current) {
			otmp := offset
			for isSpace(rune(data[otmp])) {
				otmp++
			}
			if data[otmp] == ']' {
				// special case an array offset miss
				offset = otmp
				return nil, nil
			}
			val, _, err := json.NextValue(data[offset:], scan)
			return val, err
		}
	}

	return nil, nil
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

// ListPointers lists all possible pointers from the given input.
func ListPointers(data []byte) ([]string, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("Invalid JSON")
	}
	rv := []string{""}

	scan := &json.Scanner{}
	scan.Reset()

	offset := 0
	beganLiteral := 0
	var current []string
	for {
		if offset >= len(data) {
			return rv, nil
		}
		newOp := scan.Step(scan, int(data[offset]))
		offset++

		switch newOp {
		case json.ScanBeginArray:
			current = append(current, "0")
		case json.ScanObjectKey:
			current[len(current)-1] = grokLiteral(data[beganLiteral-1 : offset-1])
		case json.ScanBeginLiteral:
			beganLiteral = offset
		case json.ScanArrayValue:
			n := mustParseInt(current[len(current)-1])
			current[len(current)-1] = strconv.Itoa(n + 1)
		case json.ScanEndArray, json.ScanEndObject:
			current = sliceToEnd(current)
		case json.ScanBeginObject:
			current = append(current, "")
		case json.ScanError:
			return nil, fmt.Errorf("Error reading JSON object at offset %v", offset)
		}

		if newOp == json.ScanBeginArray || newOp == json.ScanArrayValue ||
			newOp == json.ScanObjectKey {
			rv = append(rv, encodePointer(current))
		}
	}
}

// FindMany finds several jsonpointers in one pass through the input.
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
			current[len(current)-1] = grokLiteral(data[beganLiteral-1 : offset-1])
		case json.ScanBeginLiteral:
			beganLiteral = offset
		case json.ScanArrayValue:
			n := mustParseInt(current[len(current)-1])
			current[len(current)-1] = strconv.Itoa(n + 1)
		case json.ScanEndArray, json.ScanEndObject:
			current = sliceToEnd(current)
		case json.ScanBeginObject:
			current = append(current, "")
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

// Borrowed from stdlib (older version).  hopefully can just get it exported.
func unquoteBytes(s []byte) (t []byte, ok bool) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		s = bytes.TrimSpace(s)
		if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
			return
		}
	}
	s = s[1 : len(s)-1]

	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < len(s) {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == len(s) {
		return s, true
	}

	b := make([]byte, len(s)+2*utf8.UTFMax)
	w := copy(b, s[0:r])
	for r < len(s) {
		// Out of room?  Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= len(s) {
				return
			}
			switch s[r] {
			default:
				return
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := getu4(s[r:])
				if rr < 0 {
					return
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return b[0:w], true
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	r, err := strconv.ParseUint(string(s[2:6]), 16, 64)
	if err != nil {
		return -1
	}
	return rune(r)
}
