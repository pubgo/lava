// Package lava re-exports the public API from pkg/lava for backward compatibility.
//
// Deprecated: import github.com/pubgo/lava/v2/pkg/lava instead.
package lava

import plava "github.com/pubgo/lava/v2/pkg/lava"

type (
	RequestKind    = plava.RequestKind
	Request        = plava.Request
	RequestHeader  = plava.RequestHeader
	Response       = plava.Response
	ResponseHeader = plava.ResponseHeader
	HandlerFunc    = plava.HandlerFunc
	Middlewares    = plava.Middlewares
	Middleware     = plava.Middleware
	MiddlewareWrap = plava.MiddlewareWrap
	GrpcProxyCfg   = plava.GrpcProxyCfg
	GrpcProxy      = plava.GrpcProxy
	GrpcHttpRouter = plava.GrpcHttpRouter
	GrpcRouter     = plava.GrpcRouter
	HttpRouter     = plava.HttpRouter
	Closer         = plava.Closer
	Listener       = plava.Listener
	Validator      = plava.Validator
)

const (
	RequestKindHttp = plava.RequestKindHttp
	RequestKindGrpc = plava.RequestKindGrpc
	RequestKindZrpc = plava.RequestKindZrpc
)

var (
	WithMiddleware = plava.WithMiddleware
	Chain          = plava.Chain
)
