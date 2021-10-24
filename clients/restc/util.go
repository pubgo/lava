package restc

import (
	"github.com/pubgo/lava/types"
	"github.com/pubgo/x/byteutil"
)

func convertHeader(request interface{ VisitAll(func(key, value []byte)) }) types.Header {
	var h = types.HeaderGet()
	request.VisitAll(func(key, value []byte) {
		h.Add(byteutil.ToStr(key), byteutil.ToStr(value))
	})
	return h
}
