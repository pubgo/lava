package lava

import (
	"context"

	"github.com/rs/xid"
)

var reqCtxKey = xid.New().String()

func CreateCtxWithReq(ctx context.Context, header *RequestHeader) context.Context {
	return context.WithValue(ctx, reqCtxKey, header)
}

func GetReqHeader(ctx context.Context) *RequestHeader {
	val, ok := ctx.Value(reqCtxKey).(*RequestHeader)
	if ok {
		return val
	}
	return nil
}

var rspCtxKey = xid.New().String()

func CreateCtxWithRsp(ctx context.Context, header *ResponseHeader) context.Context {
	return context.WithValue(ctx, rspCtxKey, header)
}

func GetRspHeader(ctx context.Context) *ResponseHeader {
	val, ok := ctx.Value(rspCtxKey).(*ResponseHeader)
	if ok {
		return val
	}
	return val
}

var reqIdKey = xid.New().String()

func CreateCtxWithReqID(ctx context.Context, reqId string) context.Context {
	return context.WithValue(ctx, reqIdKey, reqId)
}

func GetReqID(ctx context.Context) string {
	var reqId, ok = ctx.Value(reqIdKey).(string)
	if ok {
		return reqId
	}
	return ""
}
