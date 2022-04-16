package abc

import (
	"google.golang.org/grpc/metadata"
)

type Header = metadata.MD

func HeaderGet(h Header, name string) string {
	val := h.Get(name)
	if len(val) != 0 && val[0] != "" {
		return val[0]
	}
	return ""
}
