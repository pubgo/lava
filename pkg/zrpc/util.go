package zrpc

import (
	"context"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/rs/xid"

	"github.com/pubgo/lava/v2/core/lavacontexts"
	"github.com/pubgo/lava/v2/lava"
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

func requestHeaderFromContext(ctx context.Context) *lava.RequestHeader {
	if header := contextRequestHeader(ctx); header != nil {
		dup := new(lava.RequestHeader)
		copyRequestHeader(dup, header)
		return dup
	}

	return new(lava.RequestHeader)
}

func requestHeaderFromNATS(subject string, header nats.Header) *lava.RequestHeader {
	reqHeader := new(lava.RequestHeader)
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

func responseHeaderFromNATS(header nats.Header) *lava.ResponseHeader {
	rspHeader := new(lava.ResponseHeader)
	for key, values := range header {
		for _, value := range values {
			rspHeader.Add(key, value)
		}
	}

	return rspHeader
}

func requestHeaderToNATS(header *lava.RequestHeader) nats.Header {
	result := nats.Header{}
	if header == nil {
		return result
	}

	for key, value := range header.All() {
		result.Add(string(key), string(value))
	}

	return result
}

func responseHeaderToNATS(header *lava.ResponseHeader) nats.Header {
	result := nats.Header{}
	if header == nil {
		return result
	}

	for key, value := range header.All() {
		result.Add(string(key), string(value))
	}

	return result
}

func copyRequestHeader(dst, src *lava.RequestHeader) {
	if dst == nil || src == nil {
		return
	}

	for key, value := range src.All() {
		dst.Add(string(key), string(value))
	}
	dst.SetMethodBytes(src.Method())
	dst.SetRequestURIBytes(src.RequestURI())
	dst.SetContentTypeBytes(src.ContentType())
}

func copyResponseHeader(dst, src *lava.ResponseHeader) {
	if dst == nil || src == nil {
		return
	}

	for key, value := range src.All() {
		dst.Add(string(key), string(value))
	}
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

func contextRequestHeader(ctx context.Context) (header *lava.RequestHeader) {
	defer func() {
		if recover() != nil {
			header = nil
		}
	}()

	return lavacontexts.ReqHeader(ctx)
}

func contextResponseHeader(ctx context.Context) (header *lava.ResponseHeader) {
	defer func() {
		if recover() != nil {
			header = nil
		}
	}()

	return lavacontexts.RspHeader(ctx)
}
