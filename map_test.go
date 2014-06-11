package jsonpointer

import (
	"io/ioutil"
	"reflect"
	"strings"
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
	{"/foo/99", nil},
	{"/foo/0/3", nil},
	{"/foo/0", "bar"},
	{"/foo", []interface{}{"bar", "baz"}},
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
	{"", obj},
}

func unmarshalObjSrc() {
	err := json.Unmarshal([]byte(objSrc), &obj)
	if err != nil {
		panic(err)
	}
}

func TestPaths(t *testing.T) {
	unmarshalObjSrc()
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

func TestSet(t *testing.T) {
	unmarshalObjSrc()
	for _, test := range tests {
		if test.path == "" {
			continue
		}

		err := Set(obj, test.path, "testing set")
		if test.exp == nil {
			if err == nil {
				t.Errorf("expected error for %v", test.path)
				t.Fail()
			}
			continue
		} else if err != nil {
			t.Errorf("On %v", test.path)
			t.Fail()
		}

		got := Get(obj, test.path)
		if got != "testing set" {
			t.Errorf("On %v, expected `testing set`, got %+v (%T)",
				test.path, got, got)
			t.Fail()
		} else {
			t.Logf("Success - got %v for %v", got, test.path)
		}
	}
}

func TestDelete(t *testing.T) {
	unmarshalObjSrc()
	for _, test := range tests {
		if test.path == "" {
			continue
		}
		err := Delete(obj, test.path)
		if strings.HasPrefix(test.path, "/foo/") {
			if err == nil {
				t.Errorf("expected error on %v, can't be idempotent", test.path)
				t.Fail()
			}
			continue
		} else if err != nil {
			t.Errorf("On %v", test.path)
			t.Fail()
		}
		got := Get(obj, test.path)
		if got != nil {
			t.Errorf("On %v, expected `nil`, got %+v (%T)", test.path, got, got)
			t.Fail()
		} else {
			t.Logf("Success - got %v for %v", got, test.path)
		}
	}
}

func TestDeleteAny(t *testing.T) {
	unmarshalObjSrc()
	for _, test := range tests {
		if test.path == "" {
			continue
		}
		err := DeleteAny(obj, test.path)
		if test.exp == nil {
			if err == nil {
				t.Errorf("expected error for %v", test.path)
				t.Fail()
			}
			continue
		} else if err != nil {
			t.Errorf("On %v", test.path)
			t.Fail()
		}
		got := Get(obj, test.path)
		if test.exp == "bar" {
			if got.(string) != "baz" {
				t.Errorf("On %v, expected [baz], got %+v (%T)",
					test.path, got, got)
				t.Fail()
			}
		} else if got != nil {
			t.Errorf("On %v, expected `nil`, got %+v (%T)", test.path, got, got)
			t.Fail()
		} else {
			t.Logf("Success - got %v for %v", got, test.path)
		}
	}
}

func TestDeleteAny1Len(t *testing.T) {
	m := map[string]interface{}{
		"foo": []interface{}{"bar"},
	}
	path := "/foo/0"
	err := DeleteAny(m, path)
	if err != nil {
		t.Errorf("On %v", path)
		t.Fail()
	}
	got := Get(m, path)
	if got != nil {
		t.Errorf("On %v, expected `nil`, got %+v (%T)", path, got, got)
		t.Fail()
	} else {
		t.Logf("Success - got %v for %v", got, path)
	}
}

func BenchmarkGet357(b *testing.B) {
	var obj map[string]interface{}
	json.Unmarshal([]byte(three57JSON), &obj)
	l := len(three57Ptrs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := three57Ptrs[i%l]
		Get(obj, path)
	}
}

func BenchmarkGetPools(b *testing.B) {
	var obj map[string]interface{}
	json.Unmarshal([]byte(poolsJSON), &obj)
	l := len(poolsPtrs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := poolsPtrs[i%l]
		Get(obj, path)
	}
}

func BenchmarkGetSample(b *testing.B) {
	var obj map[string]interface{}
	json.Unmarshal([]byte(serieslysampleJSON), &obj)
	l := len(serieslysamplePtrs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := serieslysamplePtrs[i%l]
		Get(obj, path)
	}
}

func BenchmarkGetCode(b *testing.B) {
	var obj map[string]interface{}
	json.Unmarshal([]byte(codeJSON), &obj)
	l := len(codePtrs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := codePtrs[i%l]
		Get(obj, path)
	}
}

func BenchmarkSet357(b *testing.B) {
	var obj map[string]interface{}
	json.Unmarshal([]byte(three57JSON), &obj)
	l := len(three57Ptrs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := three57Ptrs[i%l]
		if path == "" {
			continue
		}
		Set(obj, path, "bench")
	}
}

func BenchmarkSetPools(b *testing.B) {
	var obj map[string]interface{}
	json.Unmarshal([]byte(poolsJSON), &obj)
	l := len(poolsPtrs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := poolsPtrs[i%l]
		if path == "" {
			continue
		}
		Set(obj, path, "bench")
	}
}

func BenchmarkSetSample(b *testing.B) {
	var obj map[string]interface{}
	json.Unmarshal([]byte(serieslysampleJSON), &obj)
	l := len(serieslysamplePtrs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := serieslysamplePtrs[i%l]
		if path == "" {
			continue
		}
		Set(obj, path, "bench")
	}
}

func BenchmarkSetCode(b *testing.B) {
	var obj map[string]interface{}
	json.Unmarshal([]byte(codeJSON), &obj)
	l := len(codePtrs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := codePtrs[i%l]
		if path == "" {
			continue
		}
		Set(obj, path, "bench")
	}
}

func BenchmarkDelete(b *testing.B) {
	unmarshalObjSrc()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			if test.path == "" {
				continue
			}
			Delete(obj, test.path)
		}
	}
}

func BenchmarkDeleteAny(b *testing.B) {
	unmarshalObjSrc()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			if test.path == "" {
				continue
			}
			DeleteAny(obj, test.path)
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

func TestIndexNotFoundSameAsPropertyNotFound(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/357.json")
	if err != nil {
		t.Fatalf("Error beer-sample brewery 357 data: %v", err)
	}

	expectedResult, expectedError := Find(data, "/doesNotExist")

	missingVals := []string{
		"/address/0",
		"/address/1",
		"/address2/1",
		"/address2/2",
		"/address3/0",
		"/address3/1",
	}

	for _, a := range missingVals {
		found, err := Find(data, a)

		if !reflect.DeepEqual(err, expectedError) {
			t.Errorf("Expected %v at %v, got %v", expectedError, a, err)
		}
		if !reflect.DeepEqual(expectedResult, found) {
			t.Errorf("Expected %v at %v, got %v", expectedResult, a, found)
		}
	}
}

const bug822src = `{
      "foo": ["bar", "baz"],
      "": 0,
      "a/b": 1,
      "c%d": 2,
      "e^f": 3,
      "g|h": 4,
      "i\\j": 5,
      "k\"l": 6,
      "k2": {},
      " ": 7,
      "m~n": 8,
      "g/n/r": "has slash, will travel",
      "g": { "n": {"r": "where's tito?"}},
      "h": {}
}`

func TestListEmptyObjectPanic822(t *testing.T) {
	ptrs, err := ListPointers([]byte(bug822src))
	if err != nil {
		t.Fatalf("Error parsing: %v", err)
	}
	t.Logf("Got pointers: %v", ptrs)
}

func TestFindEmptyObjectPanic823(t *testing.T) {
	for _, test := range tests {
		_, err := Find([]byte(bug822src), test.path)
		if err != nil {
			t.Errorf("Error looking for %v: %v", test.path, err)
		}
	}
}
