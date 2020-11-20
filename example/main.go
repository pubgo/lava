package main

import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/example/ctl_entry"
	"github.com/pubgo/golug/example/entry"
	"github.com/pubgo/xerror"
)

func main() {
	xerror.Exit(golug.Init())
	xerror.Exit(golug.Run(
		entry.GetEntry(),
		ctl_entry.GetEntry(),
	))
}
