package lava

import (
	"context"

	"github.com/rs/xid"
)

var reqCtxKey = xid.New().String()
var rspCtxKey = xid.New().String()

func CtxWithReq(ctx context.Context, header *RequestHeader) context.Context {
	return context.WithValue(ctx, reqCtxKey, header)
}

func CtxWithRsp(ctx context.Context, header *ResponseHeader) context.Context {
	return context.WithValue(ctx, rspCtxKey, header)
}

func GetReqHeader(ctx context.Context) *RequestHeader {
	val, ok := ctx.Value(reqCtxKey).(*RequestHeader)
	if ok {
		return val
	}
	return nil
}

func GetRspHeader(ctx context.Context) *ResponseHeader {
	val, ok := ctx.Value(rspCtxKey).(*ResponseHeader)
	if ok {
		return val
	}
	return val
}
