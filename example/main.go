package main

import (
	"github.com/pubgo/lug"
	"github.com/pubgo/lug/example/ctl_entry"
	"github.com/pubgo/lug/example/grpc_entry"
	"github.com/pubgo/lug/example/rest_entry"
	"github.com/pubgo/lug/example/task_entry"
)

func main() {
	lug.Run(
		task_entry.GetEntry(),
		rest_entry.GetEntry(),
		ctl_entry.GetEntry(),
		grpc_entry.GetEntry(),
	)
}
