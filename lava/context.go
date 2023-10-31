package lava

import (
	"context"

	"github.com/rs/xid"

	pbv1 "github.com/pubgo/lava/pkg/proto/lava"
)

type ctxKey string

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
var reqServerInfoKey = ctxKey(xid.New().String())

func CreateCtxWithClientInfo(ctx context.Context, info *pbv1.ServiceInfo) context.Context {
	return context.WithValue(ctx, reqClientInfoKey, info)
}

func CreateCtxWithServerInfo(ctx context.Context, info *pbv1.ServiceInfo) context.Context {
	return context.WithValue(ctx, reqServerInfoKey, info)
}

func GetClientInfo(ctx context.Context) *pbv1.ServiceInfo {
	info, ok := ctx.Value(reqClientInfoKey).(*pbv1.ServiceInfo)
	if ok {
		return info
	}
	return nil
}

func GetServerInfo(ctx context.Context) *pbv1.ServiceInfo {
	info, ok := ctx.Value(reqServerInfoKey).(*pbv1.ServiceInfo)
	if ok {
		return info
	}
	return nil
}

var reqHeader = ctxKey(xid.New().String())
var rspHeader = ctxKey(xid.New().String())

func CreateReqHeader(ctx context.Context, header *RequestHeader) context.Context {
	return context.WithValue(ctx, reqHeader, header)
}

func ReqHeader(ctx context.Context) *RequestHeader {
	return ctx.Value(reqHeader).(*RequestHeader)
}

func CreateRspHeader(ctx context.Context, header *ResponseHeader) context.Context {
	return context.WithValue(ctx, rspHeader, header)
}

func RspHeader(ctx context.Context) *ResponseHeader {
	return ctx.Value(rspHeader).(*ResponseHeader)
}
