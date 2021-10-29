package httpx

import (
	"net/http"
	"strings"
)

func IsWebsocket(h http.Header) bool {
	if strings.Contains(strings.ToLower(h.Get("Connection")), "upgrade") &&
		strings.EqualFold(h.Get("Upgrade"), "websocket") {
		return true
	}
	return false
}
