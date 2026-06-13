package zrpc

import (
	"context"
	"fmt"
	"io"
	"sync"
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

// ServerStream is a bidirectional stream used by zrpc streaming handlers.
type ServerStream struct {
	nc          *nats.Conn
	reqSub      *nats.Subscription
	reqHeader   *lava.RequestHeader
	rspHeader   *lava.ResponseHeader
	requestSubj string
	responseSub string
	ctx         context.Context

	mu        sync.Mutex
	closed    bool
	sendClose bool
}

// RegisterStream binds one zrpc streaming method on subject + queue group.
func RegisterStream(
	s *Server,
	subject, queue string,
	handler func(context.Context, *ServerStream) error,
) error {
	sub, err := s.nc.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		HandleStream(msg, s.nc, subject, s.mw, handler)
	})
	if err != nil {
		return err
	}

	s.subs = append(s.subs, sub)
	return nil
}

// HandleStream handles one streaming session.
func HandleStream(
	msg *nats.Msg,
	nc *nats.Conn,
	subject string,
	middlewares []lava.Middleware,
	handler func(context.Context, *ServerStream) error,
) {
	if msg == nil || msg.Reply == "" || !isStreamMessage(msg) || msg.Header.Get(HeaderStreamFrame) != streamFrameOpen {
		ReplyError(msg, CodeInvalidArgument, "invalid stream open frame")
		return
	}

	reqSubject := msg.Header.Get(HeaderStreamReqSub)
	if reqSubject == "" {
		ReplyError(msg, CodeInvalidArgument, "missing stream request subject")
		return
	}

	ctx := context.Background()
	if timeout := msg.Header.Get(HeaderTimeout); timeout != "" {
		if dur, err := time.ParseDuration(timeout); err == nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, dur)
			defer cancel()
		}
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

	reqSub, err := nc.SubscribeSync(reqSubject)
	if err != nil {
		ReplyError(msg, CodeInternal, "failed to subscribe stream request")
		return
	}
	defer reqSub.Unsubscribe()

	stream := &ServerStream{
		nc:          nc,
		reqSub:      reqSub,
		reqHeader:   reqHeader,
		rspHeader:   rspHeader,
		requestSubj: reqSubject,
		responseSub: msg.Reply,
		ctx:         ctx,
	}

	ack := nats.NewMsg(msg.Reply)
	ack.Header = responseHeaderToNATS(rspHeader)
	ack.Header.Set(HeaderStream, "1")
	ack.Header.Set(HeaderStreamFrame, streamFrameAck)
	if err = nc.PublishMsg(ack); err != nil {
		return
	}

	wrappedReq := &request{
		client:      false,
		stream:      true,
		subject:     subject,
		service:     serviceFromSubject(subject),
		contentType: string(reqHeader.ContentType()),
		header:      reqHeader,
		payload:     nil,
	}

	inner := func(ctx context.Context, _ lava.Request) (lava.Response, error) {
		if err = handler(ctx, stream); err != nil {
			return nil, err
		}

		return &response{header: rspHeader, payload: nil, stream: true}, stream.CloseSend()
	}

	_, err = lava.Chain(middlewares...).Middleware(inner)(ctx, wrappedReq)
	if err != nil {
		code, text := StatusFromError(err)
		stream.replyError(code, text)
		return
	}
}

// Recv receives one protobuf message from request stream.
func (s *ServerStream) Recv(req proto.Message) error {
	for {
		msg, err := s.reqSub.NextMsgWithContext(s.ctx)
		if err != nil {
			return err
		}

		if !isStreamMessage(msg) {
			continue
		}

		switch msg.Header.Get(HeaderStreamFrame) {
		case streamFrameData:
			return proto.Unmarshal(msg.Data, req)
		case streamFrameEnd:
			return io.EOF
		case streamFrameError:
			if streamErr := ErrorFromMessage(msg); streamErr != nil {
				return streamErr
			}
			return io.EOF
		}
	}
}

// Send sends one protobuf message into response stream.
func (s *ServerStream) Send(resp proto.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed || s.sendClose {
		return io.EOF
	}

	data, err := proto.Marshal(resp)
	if err != nil {
		return err
	}

	msg := nats.NewMsg(s.responseSub)
	msg.Header = responseHeaderToNATS(s.rspHeader)
	msg.Header.Set(HeaderStream, "1")
	msg.Header.Set(HeaderStreamFrame, streamFrameData)
	msg.Data = data

	return s.nc.PublishMsg(msg)
}

// CloseSend closes response side stream.
func (s *ServerStream) CloseSend() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed || s.sendClose {
		return nil
	}

	msg := nats.NewMsg(s.responseSub)
	msg.Header = responseHeaderToNATS(s.rspHeader)
	msg.Header.Set(HeaderStream, "1")
	msg.Header.Set(HeaderStreamFrame, streamFrameEnd)
	if err := s.nc.PublishMsg(msg); err != nil {
		return err
	}

	s.sendClose = true
	return nil
}

// Close closes stream resources.
func (s *ServerStream) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	_ = s.CloseSend()

	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()

	if s.reqSub != nil {
		_ = s.reqSub.Unsubscribe()
	}

	return nil
}

func (s *ServerStream) replyError(code Code, text string) {
	msg := nats.NewMsg(s.responseSub)
	msg.Header = responseHeaderToNATS(s.rspHeader)
	msg.Header.Set(HeaderStream, "1")
	msg.Header.Set(HeaderStreamFrame, streamFrameError)
	msg.Header.Set(HeaderStatusCode, fmt.Sprintf("%d", code))
	msg.Header.Set(HeaderStatusMessage, text)
	msg.Data = []byte(text)
	_ = s.nc.PublishMsg(msg)
}

// Close unsubscribes all registered handlers.
func (s *Server) Close() {
	for _, sub := range s.subs {
		_ = sub.Unsubscribe()
	}

	s.subs = nil
}
