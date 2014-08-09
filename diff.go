package jsonpointer

import (
	"log"
	"reflect"
)

// DiffType represents the type of difference between two JSON
// structures.
type DiffType int

const (
	// MissingA designates a path that was missing from the first
	// argument of the diff.
	MissingA = DiffType(iota)
	// MissingB designates a path taht was missing from the second
	// argument of the diff.
	MissingB
	// Different designates a path that is found in both
	// arguments, but with different values.
	Different
)

var diffNames = []string{"missing a", "missing b", "different"}

func (d DiffType) String() string {
	return diffNames[d]
}

func must(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func pointerSet(j []byte) (map[string]bool, error) {
	a, err := ListPointers(j)
	if err != nil {
		return nil, err
	}
	rv := map[string]bool{}
	for _, v := range a {
		rv[v] = true
	}
	return rv, nil
}

// Diff returns the differences between two json blobs.
func Diff(a, b []byte) (map[string]DiffType, error) {
	amap, err := pointerSet(a)
	if err != nil {
		return nil, err
	}
	bmap, err := pointerSet(b)
	if err != nil {
		return nil, err
	}

	rv := map[string]DiffType{}

	for v := range amap {
		if v == "" {
			continue
		}
		if bmap[v] {
			var aval, bval interface{}
			must(FindDecode(a, v, &aval))
			must(FindDecode(b, v, &bval))
			if !reflect.DeepEqual(aval, bval) {
				rv[v] = Different
			}
		} else {
			rv[v] = MissingB
		}
	}

	for v := range bmap {
		if !amap[v] {
			rv[v] = MissingA
		}
	}

	return rv, nil
}
