package lava

import (
	"context"

	"github.com/rs/xid"

	lavapbv1 "github.com/pubgo/lava/pkg/proto/lava"
)

type ctxKey string

var reqCtxKey = ctxKey(xid.New().String())

func CreateCtxWithReqHeader(ctx context.Context, header *RequestHeader) context.Context {
	return context.WithValue(ctx, reqCtxKey, header)
}

func GetReqHeader(ctx context.Context) *RequestHeader {
	val, ok := ctx.Value(reqCtxKey).(*RequestHeader)
	if ok {
		return val
	}
	return nil
}

var rspCtxKey = ctxKey(xid.New().String())

func CreateCtxWithRspHeader(ctx context.Context, header *ResponseHeader) context.Context {
	return context.WithValue(ctx, rspCtxKey, header)
}

func GetRspHeader(ctx context.Context) *ResponseHeader {
	val, ok := ctx.Value(rspCtxKey).(*ResponseHeader)
	if ok {
		return val
	}
	return val
}

var reqIdKey = ctxKey(xid.New().String())

func CreateCtxWithReqID(ctx context.Context, reqId string) context.Context {
	return context.WithValue(ctx, reqIdKey, reqId)
}

func GetReqID(ctx context.Context) string {
	reqId, ok := ctx.Value(reqIdKey).(string)
	if ok {
		return reqId
	}
	return ""
}

var reqClientInfoKey = ctxKey(xid.New().String())
var reqServiceInfoKey = ctxKey(xid.New().String())

func CreateCtxWithClientInfo(ctx context.Context, info *lavapbv1.ServiceInfo) context.Context {
	return context.WithValue(ctx, reqClientInfoKey, info)
}

func CreateCtxWithServiceInfo(ctx context.Context, info *lavapbv1.ServiceInfo) context.Context {
	return context.WithValue(ctx, reqServiceInfoKey, info)
}

func GetClientInfo(ctx context.Context) *lavapbv1.ServiceInfo {
	reqId, ok := ctx.Value(reqClientInfoKey).(*lavapbv1.ServiceInfo)
	if ok {
		return reqId
	}
	return nil
}

func GetServiceInfo(ctx context.Context) *lavapbv1.ServiceInfo {
	reqId, ok := ctx.Value(reqServiceInfoKey).(*lavapbv1.ServiceInfo)
	if ok {
		return reqId
	}
	return nil
}
