package service_builder

import (
	"github.com/pubgo/lava/abc"
	"github.com/pubgo/x/byteutil"
	"google.golang.org/grpc/metadata"
)

func convertHeader(request interface{ VisitAll(func(key, value []byte)) }) abc.Header {
	var h = metadata.MD{}
	request.VisitAll(func(key, value []byte) {
		h.Set(byteutil.ToStr(key), byteutil.ToStr(value))
	})
	return h
}