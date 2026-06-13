package zrpc

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/v2/core/lavacontexts"
	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/httputil"
)

// Server registers zrpc unary handlers on NATS queue subscriptions.
type Server struct {
	nc   *nats.Conn
	subs []*nats.Subscription
	mw   []lava.Middleware
}

// NewServer creates a server bound to a NATS connection.
func NewServer(nc *nats.Conn, middlewares ...lava.Middleware) *Server {
	return &Server{nc: nc, mw: middlewares}
}

// Conn returns the underlying NATS connection.
func (s *Server) Conn() *nats.Conn {
	return s.nc
}

// RegisterUnary binds one protobuf unary method on subject + queue group.
func RegisterUnary[Req, Resp proto.Message](
	s *Server,
	subject, queue string,
	newReq func() Req,
	handler func(context.Context, Req) (Resp, error),
) error {
	sub, err := s.nc.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		HandleUnary(msg, subject, s.mw, newReq, handler)
	})
	if err != nil {
		return err
	}

	s.subs = append(s.subs, sub)
	return nil
}

// HandleUnary decodes a request, applies middlewares, calls the handler, and responds.
func HandleUnary[Req, Resp proto.Message](
	msg *nats.Msg,
	subject string,
	middlewares []lava.Middleware,
	newReq func() Req,
	handler func(context.Context, Req) (Resp, error),
) {
	ctx := context.Background()
	if timeout := msg.Header.Get(HeaderTimeout); timeout != "" {
		if dur, err := time.ParseDuration(timeout); err == nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, dur)
			defer cancel()
		}
	}

	req := newReq()
	if err := proto.Unmarshal(msg.Data, req); err != nil {
		ReplyError(msg, CodeInvalidArgument, "decode failed")
		return
	}

	reqHeader := requestHeaderFromNATS(subject, msg.Header)
	rspHeader := new(lava.ResponseHeader)
	reqID := firstNotEmpty(
		msg.Header.Get(httputil.HeaderXRequestID),
		string(reqHeader.Peek(httputil.HeaderXRequestID)),
		newRequestID(),
	)
	reqHeader.Set(httputil.HeaderXRequestID, reqID)
	reqHeader.Set(httputil.HeaderXRequestOperation, subject)

	ctx = lavacontexts.CreateCtxWithReqID(ctx, reqID)
	ctx = lavacontexts.CreateReqHeader(ctx, reqHeader)
	ctx = lavacontexts.CreateRspHeader(ctx, rspHeader)

	wrappedReq := &request{
		client:      false,
		subject:     subject,
		service:     serviceFromSubject(subject),
		contentType: string(reqHeader.ContentType()),
		header:      reqHeader,
		payload:     req,
	}

	inner := func(ctx context.Context, _ lava.Request) (lava.Response, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, err
		}

		return &response{header: rspHeader, payload: resp}, nil
	}

	wrappedResp, err := lava.Chain(middlewares...).Middleware(inner)(ctx, wrappedReq)
	if err != nil {
		code, text := StatusFromError(err)
		ReplyErrorWithHeader(msg, responseHeaderToNATS(rspHeader), code, text)
		return
	}

	typedResp, ok := wrappedResp.Payload().(proto.Message)
	if !ok || typedResp == nil {
		ReplyErrorWithHeader(msg, responseHeaderToNATS(wrappedResp.Header()), CodeInternal, "invalid response payload")
		return
	}

	data, err := proto.Marshal(typedResp)
	if err != nil {
		ReplyErrorWithHeader(msg, responseHeaderToNATS(wrappedResp.Header()), CodeInternal, "encode failed")
		return
	}

	_ = msg.RespondMsg(&nats.Msg{Header: responseHeaderToNATS(wrappedResp.Header()), Data: data})
}

// Close unsubscribes all registered handlers.
func (s *Server) Close() {
	for _, sub := range s.subs {
		_ = sub.Unsubscribe()
	}

	s.subs = nil
}
