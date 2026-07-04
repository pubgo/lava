package zrpc

import (
	"context"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/rs/xid"

	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/lavacontexts"
)

const (
	Name               = "zrpc"
	DefaultContentType = "application/protobuf"
	HeaderTimeout      = "Timeout" // propagated to server; also used for stream session deadline
	MethodNATS         = "NATS"
	HeaderStream       = "Zrpc-Stream"
	HeaderStreamFrame  = "Zrpc-Stream-Frame"
	HeaderStreamReqSub = "Zrpc-Stream-Req-Subject"

	streamFrameOpen  = "open"
	streamFrameAck   = "ack"
	streamFrameData  = "data"
	streamFrameEnd   = "end"
	streamFrameError = "error"
)

func newRequestID() string {
	return xid.New().String()
}

func firstNotEmpty(vals ...string) string {
	for _, val := range vals {
		if val != "" {
			return val
		}
	}

	return ""
}

func serviceFromSubject(subject string) string {
	service := subject
	if idx := strings.Index(subject, "/"); idx >= 0 {
		service = subject[:idx]
	}

	return strings.TrimPrefix(service, "svc.")
}

func requestHeaderFromContext(ctx context.Context) lava.RequestHeader {
	if header := contextRequestHeader(ctx); header != nil {
		dup := httputil.NewRequestHeader()
		copyRequestHeader(dup, header)
		return dup
	}

	return httputil.NewRequestHeader()
}

func requestHeaderFromNATS(subject string, header nats.Header) lava.RequestHeader {
	reqHeader := httputil.NewRequestHeader()
	reqHeader.SetMethod(MethodNATS)
	reqHeader.SetRequestURI(subject)
	reqHeader.SetContentType(DefaultContentType)
	for key, values := range header {
		for _, value := range values {
			reqHeader.Add(key, value)
		}
	}

	return reqHeader
}

func responseHeaderFromNATS(header nats.Header) lava.ResponseHeader {
	rspHeader := httputil.NewResponseHeader()
	for key, values := range header {
		for _, value := range values {
			rspHeader.Add(key, value)
		}
	}

	return rspHeader
}

func requestHeaderToNATS(header lava.RequestHeader) nats.Header {
	result := nats.Header{}
	if header == nil {
		return result
	}

	header.VisitAll(func(key, value []byte) {
		result.Add(string(key), string(value))
	})

	return result
}

func responseHeaderToNATS(header lava.ResponseHeader) nats.Header {
	result := nats.Header{}
	if header == nil {
		return result
	}

	header.VisitAll(func(key, value []byte) {
		result.Add(string(key), string(value))
	})

	return result
}

func copyRequestHeader(dst, src lava.RequestHeader) {
	if dst == nil || src == nil {
		return
	}

	src.VisitAll(func(key, value []byte) {
		dst.Add(string(key), string(value))
	})
	dst.SetMethodBytes(src.Method())
	dst.SetRequestURIBytes(src.RequestURI())
	dst.SetContentTypeBytes(src.ContentType())
}

func copyResponseHeader(dst, src lava.ResponseHeader) {
	if dst == nil || src == nil {
		return
	}

	src.VisitAll(func(key, value []byte) {
		dst.Add(string(key), string(value))
	})
}

func cloneHeader(header nats.Header) nats.Header {
	cloned := nats.Header{}
	for key, values := range header {
		for _, value := range values {
			cloned.Add(key, value)
		}
	}

	return cloned
}

// contextRequestHeader 返回 context 中的请求 Header，不存在时返回 nil。
// lavacontexts.ReqHeader 已保证不会 panic，这里直接透传。
func contextRequestHeader(ctx context.Context) lava.RequestHeader {
	return lavacontexts.ReqHeader(ctx)
}

// contextResponseHeader 返回 context 中的响应 Header，不存在时返回 nil。
// lavacontexts.RspHeader 已保证不会 panic，这里直接透传。
func contextResponseHeader(ctx context.Context) lava.ResponseHeader {
	return lavacontexts.RspHeader(ctx)
}
