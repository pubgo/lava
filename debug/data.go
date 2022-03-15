package debug

import (
	"net/http"
	"sync"
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
