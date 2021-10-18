package main

import (
	"fmt"
	"net/url"
	_ "unsafe"

	_ "github.com/gin-gonic/gin/binding"
	"github.com/gorilla/schema"
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
)

//go:linkname mapFormByTag github.com/gin-gonic/gin/binding.mapFormByTag
func mapFormByTag(ptr interface{}, form map[string][]string, tag string) error

type t1 struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	var n t1
	xerror.Panic(mapFormByTag(&n, map[string][]string{"name": {"hello"}}, "json"))
	q.Q(n)

	n = t1{}
	var decoder = schema.NewDecoder()
	decoder.SetAliasTag("json")
	xerror.Panic(decoder.Decode(&n, map[string][]string{"name": {"hello11"}}))
	q.Q(n)

	var rr, err = url.ParseQuery("a=1&a=2")
	xerror.Panic(err)
	fmt.Println(rr)
}
