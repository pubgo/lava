package main

import (
	"fmt"
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/example/entry"
	"github.com/pubgo/xerror"
)

func main() {
	defer xerror.Resp(func(err xerror.XErr) {
		fmt.Println(err.Println())
	})
	xerror.Panic(golug.Init())
	xerror.Panic(golug.Run(
		entry.GetEntry(),
	))
}
