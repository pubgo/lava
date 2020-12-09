package main

import (
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/cmd/ossync/rsync"
	"github.com/pubgo/xerror"
)

func main() {
	xerror.Exit(golug.Init())
	xerror.Exit(golug.Run(rsync.GetEntry()))
}
