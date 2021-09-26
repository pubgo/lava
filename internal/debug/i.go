package debug

import (
	"net/http"
	"sync"
	_ "unsafe"
)

type muxEntry struct {
	h       http.Handler
	pattern string
}

type ServeMux struct {
	mu    sync.RWMutex
	m     map[string]muxEntry
	es    []muxEntry // slice of entries sorted from longest to shortest.
	hosts bool       // whether any patterns contain hostnames
}

//go:linkname serveMux net/http.DefaultServeMux
var serveMux *ServeMux
