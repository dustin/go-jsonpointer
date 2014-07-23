package jsonpointer

import (
	"reflect"
	"testing"
)

type address struct {
	Street string `json:"street"`
	Zip    string
}

type person struct {
	Name               string `json:"name,omitempty"`
	Twitter            string
	Aliases            []string   `json:"aliases"`
	Addresses          []*address `json:"addresses"`
	NameTildeContained string     `json:"name~contained"`
	NameSlashContained string     `json:"name/contained"`
}

var input = &person{
	Name:    "marty",
	Twitter: "mschoch",
	Aliases: []string{
		"jabroni",
		"beer",
	},
	Addresses: []*address{
		&address{
			Street: "123 Sesame St.",
			Zip:    "99099",
		},
	},
	NameTildeContained: "yessir",
	NameSlashContained: "nosir",
}

func TestReflectListPointers(t *testing.T) {
	pointers, err := ReflectListPointers(input)
	if err != nil {
		t.Fatal(err)
	}
	expect := []string{"", "/name", "/Twitter", "/aliases", "/aliases/0", "/aliases/1", "/addresses", "/addresses/0", "/addresses/0/street", "/addresses/0/Zip", "/name~0contained", "/name~1contained"}
	if !reflect.DeepEqual(pointers, expect) {
		t.Fatalf("expected %#v, got %#v", expect, pointers)
	}
}

func TestReflectNonObjectOrSlice(t *testing.T) {
	got := Reflect(36, "/test")
	if got != nil {
		t.Errorf("expected nil, got %#v", got)
	}
}

func TestReflect(t *testing.T) {

	tests := []struct {
		path string
		exp  interface{}
	}{
		{
			path: "",
			exp:  input,
		},
		{
			path: "/name",
			exp:  "marty",
		},
		{
			path: "/Name",
			exp:  "marty",
		},
		{
			path: "/Twitter",
			exp:  "mschoch",
		},
		{
			path: "/aliases/0",
			exp:  "jabroni",
		},
		{
			path: "/Aliases/0",
			exp:  "jabroni",
		},
		{
			path: "/addresses/0/street",
			exp:  "123 Sesame St.",
		},
		{
			path: "/addresses/4/street",
			exp:  nil,
		},
		{
			path: "/doesntexist",
			exp:  nil,
		},
		{
			path: "/does/not/exit",
			exp:  nil,
		},
		{
			path: "/doesntexist/7",
			exp:  nil,
		},
		{
			path: "/name~0contained",
			exp:  "yessir",
		},
		{
			path: "/name~1contained",
			exp:  "nosir",
		},
	}

	for _, test := range tests {
		output := Reflect(input, test.path)
		if !reflect.DeepEqual(output, test.exp) {
			t.Errorf("Expected %#v, got %#v", test.exp, output)
		}
	}
}
