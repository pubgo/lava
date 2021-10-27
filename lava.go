package lava

import (
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/internal/runtime"
)

func Run(description string, entries ...entry.Entry) {
	runtime.Run(description, entries...)
}
