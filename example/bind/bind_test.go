package main

import (
	"github.com/gorilla/schema"

	"testing"
)

func BenchmarkMapFormByTag(b *testing.B) {
	var n t1
	var data = map[string][]string{"name": {"hello"}, "age": {"1"}}
	for i := 0; i < b.N; i++ {
		if err := mapFormByTag(&n, data, "json"); err != nil {
			b.Fatal(err)
		}

	}
}

func BenchmarkSchema(b *testing.B) {
	var n t1
	var data = map[string][]string{"name": {"hello111"}, "age": {"1"}}
	var decoder = schema.NewDecoder()
	decoder.SetAliasTag("json")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := decoder.Decode(&n, data); err != nil {
			b.Fatal(err)
		}
	}
}
