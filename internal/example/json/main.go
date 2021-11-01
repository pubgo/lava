package main

import (
	"encoding/json"
	"fmt"

	"github.com/pubgo/xerror"
)

// https://github.com/goccy/go-json

type a struct {
	Hello  string `json:"hello"`
	Hello1 string `json:"Hello1"`
}

func main() {
	var d a
	xerror.Panic(json.Unmarshal([]byte(`{"Hello":"a","hello1":"b"}`), &d))
	fmt.Printf("%#v\n", d)
}
