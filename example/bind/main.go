package main

import (
	_ "github.com/gin-gonic/gin/binding"
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	_ "unsafe"
)

//go:linkname mapFormByTag github.com/gin-gonic/gin/binding.mapFormByTag
func mapFormByTag(ptr interface{}, form map[string][]string, tag string) error

type t1 struct {
	Name string `json:"name"`
}

func main() {
	var n t1
	xerror.Panic(mapFormByTag(&n, map[string][]string{"name": {"hello"}}, "json"))
	q.Q(n)
}
