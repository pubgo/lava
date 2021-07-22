package main

import (
	"fmt"

	"github.com/pubgo/lug"
	"github.com/pubgo/lug/example/ctl_entry"
	"github.com/pubgo/lug/example/grpc_entry"
	"github.com/pubgo/lug/example/rest_entry"
	"github.com/pubgo/lug/example/task_entry"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
)

func main() {
	lug.AfterStart(func() {
		fmt.Println(dix.Graph())
	})

	xerror.Exit(lug.Run(
		"lug example 测试",
		task_entry.GetEntry(),
		rest_entry.GetEntry(),
		ctl_entry.GetEntry(),
		grpc_entry.GetEntry(),
	))
}
