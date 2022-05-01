package httpx

import (
	"strings"

	"github.com/valyala/fasthttp"
)

func IsWebsocket(h *fasthttp.RequestHeader) bool {
	if strings.Contains(strings.ToLower(string(h.Peek("Connection"))), "upgrade") &&
		strings.EqualFold(string(h.Peek("Upgrade")), "websocket") {
		return true
	}
	return false
}
