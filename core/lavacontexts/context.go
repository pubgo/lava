package lavacontexts

import (
	"context"

	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/proto/lavapbv1"
	"github.com/rs/xid"
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

var (
	reqClientInfoKey = ctxKey(xid.New().String())
	reqServerInfoKey = ctxKey(xid.New().String())
)

func CreateCtxWithClientInfo(ctx context.Context, info *lavapbv1.ServiceInfo) context.Context {
	return context.WithValue(ctx, reqClientInfoKey, info)
}

func CreateCtxWithServerInfo(ctx context.Context, info *lavapbv1.ServiceInfo) context.Context {
	return context.WithValue(ctx, reqServerInfoKey, info)
}

func GetClientInfo(ctx context.Context) *lavapbv1.ServiceInfo {
	info, ok := ctx.Value(reqClientInfoKey).(*lavapbv1.ServiceInfo)
	if ok {
		return info
	}
	return nil
}

func GetServerInfo(ctx context.Context) *lavapbv1.ServiceInfo {
	info, ok := ctx.Value(reqServerInfoKey).(*lavapbv1.ServiceInfo)
	if ok {
		return info
	}
	return nil
}

var (
	reqHeader = ctxKey(xid.New().String())
	rspHeader = ctxKey(xid.New().String())
)

func CreateReqHeader(ctx context.Context, header *lava.RequestHeader) context.Context {
	return context.WithValue(ctx, reqHeader, header)
}

func ReqHeader(ctx context.Context) *lava.RequestHeader {
	return ctx.Value(reqHeader).(*lava.RequestHeader)
}

func CreateRspHeader(ctx context.Context, header *lava.ResponseHeader) context.Context {
	return context.WithValue(ctx, rspHeader, header)
}

func RspHeader(ctx context.Context) *lava.ResponseHeader {
	return ctx.Value(rspHeader).(*lava.ResponseHeader)
}
