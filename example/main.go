package main

import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/example/ctl_entry"
	"github.com/pubgo/golug/example/grpc_entry"
	"github.com/pubgo/golug/example/rest_entry"
	"github.com/pubgo/golug/example/task_entry"
)

func main() {
	golug.Run(
		task_entry.GetEntry(),
		rest_entry.GetEntry(),
		ctl_entry.GetEntry(),
		grpc_entry.GetEntry(),
	)
}
