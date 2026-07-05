// Package lavacontexts 提供基于 context.Context 的请求级数据传递工具。
//
// 它在 context 中存取以下几类与单次 RPC/HTTP 请求绑定的数据：
//   - 请求 ID（reqId）：用于全链路日志追踪
//   - 客户端/服务端 ServiceInfo：调用双方的服务元信息
//   - 请求/响应 Header：lava 抽象 header 引用
//
// 所有 Getter 在 context 中不存在对应值时均安全返回零值（"" 或 nil），
// 不会 panic，调用方无需额外的 recover 保护。
package lavacontexts

import (
	"context"

	"github.com/rs/xid"

	"github.com/pubgo/lava/v2/pkg/lava"
	"github.com/pubgo/lava/v2/pkg/proto/lavapbv1"
)

type ctxKey string

var reqIdKey = ctxKey(xid.New().String())

// CreateCtxWithReqID 返回携带请求 ID 的新 context。
func CreateCtxWithReqID(ctx context.Context, reqId string) context.Context {
	return context.WithValue(ctx, reqIdKey, reqId)
}

// GetReqID 返回 context 中的请求 ID，不存在时返回空字符串。
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

// CreateCtxWithClientInfo 返回携带客户端 ServiceInfo 的新 context。
func CreateCtxWithClientInfo(ctx context.Context, info *lavapbv1.ServiceInfo) context.Context {
	return context.WithValue(ctx, reqClientInfoKey, info)
}

// CreateCtxWithServerInfo 返回携带服务端 ServiceInfo 的新 context。
func CreateCtxWithServerInfo(ctx context.Context, info *lavapbv1.ServiceInfo) context.Context {
	return context.WithValue(ctx, reqServerInfoKey, info)
}

// GetClientInfo 返回 context 中的客户端 ServiceInfo，不存在时返回 nil。
func GetClientInfo(ctx context.Context) *lavapbv1.ServiceInfo {
	info, ok := ctx.Value(reqClientInfoKey).(*lavapbv1.ServiceInfo)
	if ok {
		return info
	}
	return nil
}

// GetServerInfo 返回 context 中的服务端 ServiceInfo，不存在时返回 nil。
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

// CreateReqHeader 返回携带请求 Header 的新 context。
func CreateReqHeader(ctx context.Context, header lava.RequestHeader) context.Context {
	return context.WithValue(ctx, reqHeader, header)
}

// ReqHeader 返回 context 中的请求 Header，不存在时返回 nil。
func ReqHeader(ctx context.Context) lava.RequestHeader {
	header, ok := ctx.Value(reqHeader).(lava.RequestHeader)
	if ok {
		return header
	}
	return nil
}

// CreateRspHeader 返回携带响应 Header 的新 context。
func CreateRspHeader(ctx context.Context, header lava.ResponseHeader) context.Context {
	return context.WithValue(ctx, rspHeader, header)
}

// RspHeader 返回 context 中的响应 Header，不存在时返回 nil。
func RspHeader(ctx context.Context) lava.ResponseHeader {
	header, ok := ctx.Value(rspHeader).(lava.ResponseHeader)
	if ok {
		return header
	}
	return nil
}
