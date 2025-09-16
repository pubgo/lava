package main

import (
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/lava/v2/cmds/servefilecmd"
	"github.com/pubgo/lava/v2/core/lavabuilder"
)

func main() {
	defer recovery.Exit()

	builder := lavabuilder.New()
	builder.Provide(servefilecmd.New)

	lavabuilder.Run(builder)
}
