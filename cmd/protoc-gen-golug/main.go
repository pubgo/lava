package main

import (
	"log"

	"github.com/pubgo/xerror"
	"github.com/pubgo/xprotogen/gen"
)

func main() {
	m := gen.New("golug")
	m.Parameter(func(key, value string) {
		log.Println("params:", key, "=", value)
	})
	xerror.Exit(m.GenWithTpl(tpl))
}
