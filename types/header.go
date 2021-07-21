package types

import (
	"net/http"
	"sync"
)

type Header = http.Header

var headerPool = sync.Pool{
	New: func() interface{} {
		return make(http.Header)
	},
}

func HeaderGet() Header { return headerPool.Get().(Header) }
func HeaderPut(h Header) {
	defer headerPool.Put(h)
	for k := range h {
		delete(h, k)
	}
}
