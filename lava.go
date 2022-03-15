package lava

import (
	"github.com/pubgo/lava/internal/runtime"
	"github.com/pubgo/lava/server"
)

func Run(desc string, entries ...server.Entry) {
	runtime.Run(desc, entries...)
}
