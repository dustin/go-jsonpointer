package jsonpointer

import (
	"reflect"
	"testing"
)

var ptests = []struct {
	path string
	exp  interface{}
}{
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

func TestPointerRoot(t *testing.T) {
	got, err := Find([]byte(objSrc), "")
	if err != nil {
		t.Fatalf("Error finding root: %v", err)
	}
	if !reflect.DeepEqual([]byte(objSrc), got) {
		t.Fatalf("Error finding root, found\n%s\n, wanted\n%s",
			got, objSrc)
	}
}

func TestPointer(t *testing.T) {

	for _, test := range ptests {
		got, err := Find([]byte(objSrc), test.path)
		var val interface{}
		if err == nil {
			err = Unmarshal([]byte(got), &val)
		}
		if err != nil {
			t.Errorf("Got an error on key %v: %v", test.path, err)
			t.Fail()
		} else if !reflect.DeepEqual(val, test.exp) {
			t.Errorf("On %#v, expected %+v (%T), got %+v (%T)",
				test.path, test.exp, test.exp, val, val)
			t.Fail()
		} else {
			t.Logf("Success - got %s for %#v", got, test.path)
		}
	}
}

func BenchmarkPointer(b *testing.B) {
	obj := []byte(objSrc)
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			Find(obj, test.path)
		}
	}
}
