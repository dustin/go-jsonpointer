package jsonpointer

import (
	"io"
	"reflect"
	"testing"
)

func TestDiffNil(t *testing.T) {
	diffs, err := Diff(nil, nil)
	if err != nil {
		t.Fatalf("Expected no error on nil diff, got %v", err)
	}
	if len(diffs) != 0 {
		t.Errorf("Expected no diffs, got %v", diffs)
	}
}

func TestDiffNames(t *testing.T) {
	tests := map[DiffType]string{
		MissingA:  "missing a",
		MissingB:  "missing b",
		Different: "different",
	}

	for k, v := range tests {
		if k.String() != v {
			t.Errorf("Expected %v for %d, got %v", v, k, k)
		}
	}
}

func TestMust(t *testing.T) {
	must(nil) // no panic
	panicked := false
	func() {
		defer func() { panicked = recover() != nil }()
		must(io.EOF)
	}()
	if !panicked {
		t.Fatalf("Expected a panic, but didn't get one")
	}
}

func TestDiff(t *testing.T) {
	var (
		aFirst = `{"a": 1, "b": 3.2}`
		bFirst = `{"b":3.2,"a":1}`
		aTwo   = `{"a": 2, "b": 3.2}`
		aOnly1 = `{"a": 1}`
		aOnly3 = `{"a": 3}`
		broken = `{x}`
	)

	tests := []struct {
		name    string
		a, b    string
		exp     map[string]DiffType
		errored bool
	}{
		{"Empty", "", "", map[string]DiffType{}, false},
		{"Identity", aFirst, aFirst, map[string]DiffType{}, false},
		{"Same", aFirst, bFirst, map[string]DiffType{}, false},
		{"Other order", aFirst, bFirst, map[string]DiffType{}, false},
		{"A diff", aFirst, aTwo, map[string]DiffType{"/a": Different}, false},
		{"A diff rev", aTwo, aFirst, map[string]DiffType{"/a": Different}, false},
		{"Missing b <- 1", aFirst, aOnly1, map[string]DiffType{"/b": MissingB}, false},
		{"Missing b -> 1", aOnly1, aFirst, map[string]DiffType{"/b": MissingA}, false},
		{"Missing b <- 3", aTwo, aOnly3, map[string]DiffType{
			"/a": Different,
			"/b": MissingB,
		}, false},
		{"Missing b -> 3", aOnly3, aTwo, map[string]DiffType{
			"/a": Different,
			"/b": MissingA,
		}, false},
		{"Broken A", broken, aFirst, nil, true},
		{"Broken B", aFirst, broken, nil, true},
	}

	for _, test := range tests {
		diffs, err := Diff([]byte(test.a), []byte(test.b))
		if (err != nil) != test.errored {
			t.Errorf("Expected error=%v on %q:  %v", test.errored, test.name, err)
		}
		if err != nil {
			continue
		}
		if !reflect.DeepEqual(test.exp, diffs) {
			t.Errorf("Unexpected diff for %q: %v\nwanted %v",
				test.name, diffs, test.exp)
		}
	}

}
