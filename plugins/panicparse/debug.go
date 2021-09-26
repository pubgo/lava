package panicparse

import (
	"net/http"

	"github.com/maruel/panicparse/v2/stack/webstack"
)

func init() {
	http.HandleFunc("/debug/panicparse", webstack.SnapshotHandler)
}
