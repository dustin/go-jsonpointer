package jsonpointer

import (
	"compress/gzip"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/dustin/gojson"
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

func TestPointerMissing(t *testing.T) {
	got, err := Find([]byte(objSrc), "/missing")
	if err != nil {
		t.Fatalf("Error finding missing item: %v", err)
	}
	if got != nil {
		t.Fatalf("Expected nil looking for /missing, got %v",
			got)
	}
}

func TestManyPointers(t *testing.T) {
	pointers := []string{}
	exp := map[string]interface{}{}
	for _, test := range ptests {
		pointers = append(pointers, test.path)
		exp[test.path] = test.exp
	}

	rv, err := FindMany([]byte(objSrc), pointers)
	if err != nil {
		t.Fatalf("Error finding many: %v", err)
	}

	got := map[string]interface{}{}
	for k, v := range rv {
		var val interface{}
		err = json.Unmarshal(v, &val)
		if err != nil {
			t.Fatalf("Error unmarshaling %s: %v", v, err)
		}
		got[k] = val
	}

	if !reflect.DeepEqual(got, exp) {
		for k, v := range exp {
			if !reflect.DeepEqual(got[k], v) {
				t.Errorf("At %v, expected %#v, got %#v", k, v, got[k])
			}
		}
		t.Fail()
	}
}

func TestManyPointersMissing(t *testing.T) {
	got, err := FindMany([]byte(objSrc), []string{"/missing"})
	if err != nil {
		t.Fatalf("Error finding missing item: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Expected empty looking for many /missing, got %v",
			got)
	}

}

func TestPointer(t *testing.T) {

	for _, test := range ptests {
		got, err := Find([]byte(objSrc), test.path)
		var val interface{}
		if err == nil {
			err = json.Unmarshal([]byte(got), &val)
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

func TestPointerCoder(t *testing.T) {
	tests := map[string][]string{
		"/":        []string{""},
		"/a":       []string{"a"},
		"/a~1b":    []string{"a/b"},
		"/m~0n":    []string{"m~n"},
		"/ ":       []string{" "},
		"/g~1n~1r": []string{"g/n/r"},
		"/g/n/r":   []string{"g", "n", "r"},
	}

	for k, v := range tests {
		parsed := parsePointer(k)
		encoded := encodePointer(v)

		if k != encoded {
			t.Errorf("Expected to encode %#v as %#v, got %#v",
				v, k, encoded)
			t.Fail()
		}
		if !arreq(v, parsed) {
			t.Errorf("Expected to decode %#v as %#v, got %#v",
				k, v, parsed)
			t.Fail()
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

func BenchmarkManyPointer(b *testing.B) {
	pointers := []string{}
	for _, test := range ptests {
		pointers = append(pointers, test.path)
	}
	obj := []byte(objSrc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindMany(obj, pointers)
	}
}

var codeJSON []byte

func init() {
	f, err := os.Open("testdata/code.json.gz")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(gz)
	if err != nil {
		panic(err)
	}

	codeJSON = data
}

func BenchmarkPointerLarge(b *testing.B) {
	keys := []string{
		"/tree/kids/0/kids/0/name",
		"/tree/kids/0/name",
		"/tree/kids/0/kids/0/kids/0/kids/0/kids/0/name",
	}
	b.SetBytes(int64(len(codeJSON)))

	for i := 0; i < b.N; i++ {
		found, err := FindMany(codeJSON, keys)
		if err != nil || len(found) != 3 {
			b.Fatalf("Didn't find all the things from %v/%v",
				found, err)
		}
	}
}

func BenchmarkPointerLargeShallow(b *testing.B) {
	keys := []string{
		"/tree/kids/0/kids/0/kids/0/kids/0/kids/0/name",
	}
	b.SetBytes(int64(len(codeJSON)))

	for i := 0; i < b.N; i++ {
		found, err := FindMany(codeJSON, keys)
		if err != nil || len(found) != 1 {
			b.Fatalf("Didn't find all the things: %v/%v",
				found, err)
		}
	}
}

func BenchmarkPointerLargeMissing(b *testing.B) {
	keys := []string{
		"/this/does/not/exist",
	}
	b.SetBytes(int64(len(codeJSON)))

	for i := 0; i < b.N; i++ {
		found, err := FindMany(codeJSON, keys)
		if err != nil || len(found) != 0 {
			b.Fatalf("Didn't find all the things: %v/%v",
				found, err)
		}
	}
}

func BenchmarkPointerSlow(b *testing.B) {
	keys := []string{
		"/tree/kids/0/kids/0/kids/0/kids/0/kids/0/name",
	}
	b.SetBytes(int64(len(codeJSON)))

	for i := 0; i < b.N; i++ {
		m := map[string]interface{}{}
		err := json.Unmarshal(codeJSON, &m)
		if err != nil {
			b.Fatalf("Error parsing JSON: %v", err)
		}
		Get(m, keys[0])
	}
}
