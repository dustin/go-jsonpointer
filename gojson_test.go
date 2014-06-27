package jsonpointer

import (
	"encoding/json"
	gojson "github.com/dustin/gojson"
	"testing"
)

func BenchmarkEncJSON357(b *testing.B) {
	var val []interface{}
	b.SetBytes(int64(len(three57JSON)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(three57JSON, &val)
	}
}

func BenchmarkEncJSONPools(b *testing.B) {
	var val []interface{}
	b.SetBytes(int64(len(poolsJSON)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(poolsJSON, &val)
	}
}

func BenchmarkEncJSONSample(b *testing.B) {
	var val []interface{}
	b.SetBytes(int64(len(serieslysampleJSON)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(serieslysampleJSON, &val)
	}
}

func BenchmarkEncJSONCode(b *testing.B) {
	var val []interface{}
	b.SetBytes(int64(len(codeJSON)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(codeJSON, &val)
	}
}

func BenchmarkGoJSON357(b *testing.B) {
	var val []interface{}
	b.SetBytes(int64(len(three57JSON)))
	for i := 0; i < b.N; i++ {
		gojson.Unmarshal(three57JSON, &val)
	}
}

func BenchmarkGoJSONPools(b *testing.B) {
	var val []interface{}
	b.SetBytes(int64(len(poolsJSON)))
	for i := 0; i < b.N; i++ {
		gojson.Unmarshal(poolsJSON, &val)
	}
}

func BenchmarkGoJSONSample(b *testing.B) {
	var val []interface{}
	b.SetBytes(int64(len(serieslysampleJSON)))
	for i := 0; i < b.N; i++ {
		gojson.Unmarshal(serieslysampleJSON, &val)
	}
}

func BenchmarkGoJSONCode(b *testing.B) {
	var val []interface{}
	b.SetBytes(int64(len(codeJSON)))
	for i := 0; i < b.N; i++ {
		gojson.Unmarshal(codeJSON, &val)
	}
}
