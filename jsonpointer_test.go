package jsonpointer

import (
	"encoding/json"
	"reflect"
	"testing"
)

const objSrc = `{
  "a": 1,
  "b": {
    "c": 2
  },
  "d": {
    "e": [{"a":3}, {"b":4}, {"c":5}]
  },
  "g/n/r": "has slash, will travel"
}`

var obj = map[string]interface{}{}

var tests = []struct {
	path string
	exp  interface{}
}{
	{"/", obj},
	{"/a", 1.0},
	{"/b/c", 2.0},
	{"/d/e/0/a", 3.0},
	{"/d/e/1/b", 4.0},
	{"/d/e/2/c", 5.0},
	{"/x", nil},
	{"/a/c/x", nil},
	{"/g%2fn%2Fr", "has slash, will travel"},
}

func init() {
	err := json.Unmarshal([]byte(objSrc), &obj)
	if err != nil {
		panic(err)
	}
}

func TestPaths(t *testing.T) {
	for _, test := range tests {
		got := Get(obj, test.path)
		if !reflect.DeepEqual(got, test.exp) {
			t.Errorf("On %v, expected %+v (%T), got %+v (%T)",
				test.path, test.exp, test.exp, got, got)
			t.Fail()
		} else {
			t.Logf("Success - got %v for %v", got, test.path)
		}
	}
}

func BenchmarkPaths(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			Get(obj, test.path)
		}
	}
}
