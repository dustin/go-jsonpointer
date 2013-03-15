package jsonpointer

import (
	"reflect"
	"testing"

	"github.com/dustin/gojson"
)

const objSrc = `{
      "foo": ["bar", "baz"],
      "": 0,
      "a/b": 1,
      "c%d": 2,
      "e^f": 3,
      "g|h": 4,
      "i\\j": 5,
      "k\"l": 6,
      " ": 7,
      "m~n": 8,
      "g/n/r": "has slash, will travel",
      "g": { "n": {"r": "where's tito?"}}
}`

var obj = map[string]interface{}{}

var tests = []struct {
	path string
	exp  interface{}
}{
	{"", obj},
	{"/foo", []interface{}{"bar", "baz"}},
	{"/foo/0", "bar"},
	{"/", 0.0},
	{"/a~1b", 1.0},
	{"/c%d", 2.0},
	{"/e^f", 3.0},
	{"/g|h", 4.0},
	{"/i\\j", 5.0},
	{"/k\"l", 6.0},
	{"/ ", 7.0},
	{"/m~0n", 8.0},
	{"/g~1n~1r", "has slash, will travel"},
	{"/g/n/r", "where's tito?"},
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

func BenchmarkParseAndPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			o := map[string]interface{}{}
			err := json.Unmarshal([]byte(objSrc), &o)
			if err != nil {
				b.Fatalf("Error parsing: %v", err)
			}
			Get(o, test.path)
		}
	}
}

var bug3Data = []byte(`{"foo" : "bar"}`)

func TestFindSpaceBeforeColon(t *testing.T) {
	val, err := Find(bug3Data, "/foo")
	if err != nil {
		t.Fatalf("Failed to find /foo: %v", err)
	}
	x, ok := json.UnquoteBytes(val)
	if !ok {
		t.Fatalf("Failed to unquote json bytes from %q", val)
	}
	if string(x) != "bar" {
		t.Fatalf("Expected %q, got %q", "bar", val)
	}
}

func TestListSpaceBeforeColon(t *testing.T) {
	ptrs, err := ListPointers(bug3Data)
	if err != nil {
		t.Fatalf("Error listing pointers: %v", err)
	}
	if len(ptrs) != 2 || ptrs[0] != "" || ptrs[1] != "/foo" {
		t.Fatalf(`Expected ["", "/foo"], got %#v`, ptrs)
	}
}
