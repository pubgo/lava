package main

import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/example/ctl_entry"
	"github.com/pubgo/golug/example/grpc_entry"
	"github.com/pubgo/golug/example/rest_entry"
)

func main() {
	golug.Init()

	golug.Run(
		rest_entry.GetEntry(),
		ctl_entry.GetEntry(),
		grpc_entry.GetEntry(),
		grpc_entry.GetHttpEntry(),
	)
}
