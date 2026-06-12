package zrpc

import (
	"context"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

// Server registers zrpc unary handlers on NATS queue subscriptions.
type Server struct {
	nc   *nats.Conn
	subs []*nats.Subscription
}

// NewServer creates a server bound to a NATS connection.
func NewServer(nc *nats.Conn) *Server {
	return &Server{nc: nc}
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
		HandleUnary(msg, newReq, handler)
	})
	if err != nil {
		return err
	}

	s.subs = append(s.subs, sub)
	return nil
}

// HandleUnary decodes a request, calls handler, and responds.
func HandleUnary[Req, Resp proto.Message](
	msg *nats.Msg,
	newReq func() Req,
	handler func(context.Context, Req) (Resp, error),
) {
	req := newReq()
	if err := proto.Unmarshal(msg.Data, req); err != nil {
		ReplyError(msg, CodeInvalidArgument, "decode failed")
		return
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		code, text := StatusFromError(err)
		ReplyError(msg, code, text)
		return
	}

	data, err := proto.Marshal(resp)
	if err != nil {
		ReplyError(msg, CodeInternal, "encode failed")
		return
	}

	_ = msg.Respond(data)
}

// Close unsubscribes all registered handlers.
func (s *Server) Close() {
	for _, sub := range s.subs {
		_ = sub.Unsubscribe()
	}

	s.subs = nil
}
