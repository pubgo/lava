package main

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/example/gid/entry"
)

func main() {
	lava.Run(
		"gid service",
		entry.Gid(),
	)
}
