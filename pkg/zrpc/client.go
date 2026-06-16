package zrpc

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/v2/core/lavacontexts"
	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/httputil"
)

// Client sends zrpc unary requests through NATS.
type Client struct {
	nc          *nats.Conn
	middlewares []lava.Middleware
}

// NewClient creates a zrpc client bound to a NATS connection.
func NewClient(nc *nats.Conn, middlewares ...lava.Middleware) *Client {
	return &Client{nc: nc, middlewares: middlewares}
}

// Conn returns the underlying NATS connection.
func (c *Client) Conn() *nats.Conn {
	return c.nc
}

// CallUnary sends a unary protobuf request and decodes the protobuf response.
func (c *Client) CallUnary(ctx context.Context, subject string, timeout time.Duration, req, resp proto.Message) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if _, ok := ctx.Deadline(); !ok && timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	header := requestHeaderFromContext(ctx)
	if len(header.ContentType()) == 0 {
		header.SetContentType(DefaultContentType)
	}
	header.SetMethod(MethodNATS)
	header.SetRequestURI(subject)
	header.Set(httputil.HeaderXRequestOperation, subject)

	reqID := firstNotEmpty(
		lavacontexts.GetReqID(ctx),
		string(header.Peek(httputil.HeaderXRequestID)),
		newRequestID(),
	)
	header.Set(httputil.HeaderXRequestID, reqID)

	if deadline, ok := ctx.Deadline(); ok {
		header.Set(HeaderTimeout, time.Until(deadline).String())
	}

	ctx = lavacontexts.CreateCtxWithReqID(ctx, reqID)
	ctx = lavacontexts.CreateReqHeader(ctx, header)

	wrappedReq := &request{
		client:      true,
		subject:     subject,
		service:     serviceFromSubject(subject),
		contentType: string(header.ContentType()),
		header:      header,
		payload:     req,
	}

	handler := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		payload, ok := req.Payload().(proto.Message)
		if !ok || payload == nil {
			return nil, Errorf(CodeInternal, "invalid request payload")
		}

		data, err := proto.Marshal(payload)
		if err != nil {
			return nil, err
		}

		msg := nats.NewMsg(subject)
		msg.Data = data
		msg.Header = requestHeaderToNATS(req.Header())

		rspMsg, err := c.nc.RequestMsgWithContext(ctx, msg)
		if err != nil {
			return nil, err
		}

		if err = ErrorFromMessage(rspMsg); err != nil {
			return nil, err
		}

		if err = proto.Unmarshal(rspMsg.Data, resp); err != nil {
			return nil, err
		}

		wrappedRsp := &response{header: responseHeaderFromNATS(rspMsg.Header), payload: resp}
		copyResponseHeader(contextResponseHeader(ctx), wrappedRsp.header)
		return wrappedRsp, nil
	}

	_, err := lava.Chain(c.middlewares...).Middleware(handler)(ctx, wrappedReq)
	return err
}

// ClientStream is a bidirectional zrpc message stream.
type ClientStream struct {
	nc       *nats.Conn
	reqSubj  string
	respSubj string
	respSub  *nats.Subscription
	header   nats.Header
	ctx      context.Context
	cancel   context.CancelFunc

	mu        sync.Mutex
	sendClose bool
	closed    bool
}

// OpenStream opens a bidirectional stream bound to a zrpc subject.
func (c *Client) OpenStream(ctx context.Context, subject string, timeout time.Duration) (*ClientStream, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	header := requestHeaderFromContext(ctx)
	if len(header.ContentType()) == 0 {
		header.SetContentType(DefaultContentType)
	}
	header.SetMethod(MethodNATS)
	header.SetRequestURI(subject)
	header.Set(httputil.HeaderXRequestOperation, subject)

	reqID := firstNotEmpty(
		lavacontexts.GetReqID(ctx),
		string(header.Peek(httputil.HeaderXRequestID)),
		newRequestID(),
	)
	header.Set(httputil.HeaderXRequestID, reqID)

	if deadline, ok := ctx.Deadline(); ok {
		header.Set(HeaderTimeout, time.Until(deadline).String())
	} else if timeout > 0 {
		header.Set(HeaderTimeout, timeout.String())
	}

	ctx = lavacontexts.CreateCtxWithReqID(ctx, reqID)
	ctx = lavacontexts.CreateReqHeader(ctx, header)

	handshakeCtx := ctx
	if _, ok := ctx.Deadline(); !ok && timeout > 0 {
		var cancel context.CancelFunc
		handshakeCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	reqSubj := nats.NewInbox()
	respSubj := nats.NewInbox()

	respSub, err := c.nc.SubscribeSync(respSubj)
	if err != nil {
		return nil, err
	}

	msg := nats.NewMsg(subject)
	msg.Reply = respSubj
	msg.Header = requestHeaderToNATS(header)
	msg.Header.Set(HeaderStream, "1")
	msg.Header.Set(HeaderStreamFrame, streamFrameOpen)
	msg.Header.Set(HeaderStreamReqSub, reqSubj)

	if err = c.nc.PublishMsg(msg); err != nil {
		_ = respSub.Unsubscribe()
		return nil, err
	}

	if err = c.nc.Flush(); err != nil {
		_ = respSub.Unsubscribe()
		return nil, err
	}

	for {
		ack, ackErr := respSub.NextMsgWithContext(handshakeCtx)
		if ackErr != nil {
			_ = respSub.Unsubscribe()
			return nil, ackErr
		}

		if !isStreamMessage(ack) {
			continue
		}

		switch ack.Header.Get(HeaderStreamFrame) {
		case streamFrameAck:
			streamCtx := ctx
			var cancel context.CancelFunc
			if _, ok := ctx.Deadline(); !ok && timeout > 0 {
				streamCtx, cancel = context.WithTimeout(ctx, timeout)
			} else {
				streamCtx, cancel = context.WithCancel(ctx)
			}
			return &ClientStream{
				nc:       c.nc,
				reqSubj:  reqSubj,
				respSubj: respSubj,
				respSub:  respSub,
				header:   requestHeaderToNATS(header),
				ctx:      streamCtx,
				cancel:   cancel,
			}, nil
		case streamFrameError:
			_ = respSub.Unsubscribe()
			if err = ErrorFromMessage(ack); err != nil {
				return nil, err
			}
			return nil, io.EOF
		}
	}
}

// Send sends one protobuf message into the request stream.
func (s *ClientStream) Send(msg proto.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return io.EOF
	}

	if s.sendClose {
		return io.EOF
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	frame := nats.NewMsg(s.reqSubj)
	frame.Header = cloneHeader(s.header)
	frame.Header.Set(HeaderStream, "1")
	frame.Header.Set(HeaderStreamFrame, streamFrameData)
	frame.Data = data

	return s.nc.PublishMsg(frame)
}

// Recv receives one protobuf message from the response stream.
func (s *ClientStream) Recv(resp proto.Message) error {
	for {
		msg, err := s.respSub.NextMsgWithContext(s.ctx)
		if err != nil {
			return err
		}

		if !isStreamMessage(msg) {
			continue
		}

		switch msg.Header.Get(HeaderStreamFrame) {
		case streamFrameAck:
			continue
		case streamFrameData:
			return proto.Unmarshal(msg.Data, resp)
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

// CloseSend closes the request side of the stream.
func (s *ClientStream) CloseSend() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.closeSendLocked()
}

func (s *ClientStream) closeSendLocked() error {
	if s.closed || s.sendClose {
		return nil
	}

	frame := nats.NewMsg(s.reqSubj)
	frame.Header = cloneHeader(s.header)
	frame.Header.Set(HeaderStream, "1")
	frame.Header.Set(HeaderStreamFrame, streamFrameEnd)
	if err := s.nc.PublishMsg(frame); err != nil {
		return err
	}

	s.sendClose = true
	return nil
}

// Close closes the stream and releases subscriptions.
func (s *ClientStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	_ = s.closeSendLocked()
	s.closed = true

	if s.cancel != nil {
		s.cancel()
	}
	if s.respSub != nil {
		_ = s.respSub.Unsubscribe()
	}

	return nil
}

func isStreamMessage(msg *nats.Msg) bool {
	if msg == nil || msg.Header == nil {
		return false
	}

	return msg.Header.Get(HeaderStream) == "1"
}
