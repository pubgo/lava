package debug

import (
	"net/http"
	"sync"
	"unsafe"
)

type ServeMux struct {
	Mu    sync.RWMutex
	M     map[string]muxEntry
	Es    []muxEntry
	Hosts bool
}

type muxEntry struct {
	H       http.Handler
	Pattern string
}

func GetDefaultServeMux() *ServeMux {
	return (*ServeMux)(unsafe.Pointer(http.DefaultServeMux))
}
