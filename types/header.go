package types

import (
	"sync"

	"google.golang.org/grpc/metadata"
)

type Header = metadata.MD

var headerPool = sync.Pool{
	New: func() interface{} {
		return make(metadata.MD)
	},
}

func HeaderGet() Header { return headerPool.Get().(Header) }
func HeaderPut(h Header) {
	defer headerPool.Put(h)
	for k := range h {
		delete(h, k)
	}
}
