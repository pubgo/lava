//go:generate go-bindata -fs -pkg docs -o docs/docs.go -prefix "docs/" -ignore=docs\.go docs/...

package main

import (
	"github.com/pubgo/lug"
	"github.com/pubgo/lug/example/gid/entry"
)

func main() {
	lug.Run(
		"gid service",
		entry.Gid(),
	)
}
