package lava

import (
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/internal/runtime"
)

func Run(desc string, entries ...entry.Entry) {
	runtime.Run(desc, entries...)
}

func NewService(name string, desc string) {
}
