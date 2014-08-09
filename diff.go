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

// Diff returns the differences between two json blobs.
func Diff(a, b []byte) (map[string]DiffType, error) {
	alist, err := ListPointers(a)
	if err != nil {
		return nil, err
	}
	blist, err := ListPointers(b)
	if err != nil {
		return nil, err
	}

	amap := map[string]bool{}
	bmap := map[string]bool{}
	for _, v := range alist {
		amap[v] = true
	}
	for _, v := range blist {
		bmap[v] = true
	}

	rv := map[string]DiffType{}

	for _, v := range alist {
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

	for _, v := range blist {
		if !amap[v] {
			rv[v] = MissingA
		}
	}

	return rv, nil
}
