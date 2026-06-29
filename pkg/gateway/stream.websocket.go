package gateway

import (
	"context"
	"io"
	"net/url"

	"github.com/coder/websocket"
	"github.com/pubgo/funk/v2"
	"github.com/pubgo/funk/v2/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/v2/core/encoding/protojson"
	"github.com/pubgo/lava/v2/pkg/gateway/gatewayutils"
)

// wsEncoding selects the wire format used for a websocket stream.
type wsEncoding int

const (
	wsEncodingJSON  wsEncoding = iota // protojson over text frames
	wsEncodingProto                   // protobuf over binary frames
)

var _ grpc.ServerStream = (*streamWS)(nil)

// streamWS adapts a coder/websocket connection to a grpc.ServerStream so it can
// be driven by the unified Dispatcher.
type streamWS struct {
	conn     *websocket.Conn
	ctx      context.Context
	method   *methodWrapper
	encoding wsEncoding

	params  url.Values
	header  metadata.MD
	trailer metadata.MD

	sentHeader   bool
	paramsMerged bool
}

func (s *streamWS) SetHeader(md metadata.MD) error {
	s.header = metadata.Join(s.header, md)
	return nil
}

func (s *streamWS) SendHeader(md metadata.MD) error {
	s.header = metadata.Join(s.header, md)
	s.sentHeader = true
	return nil
}

func (s *streamWS) SetTrailer(md metadata.MD) {
	s.trailer = metadata.Join(s.trailer, md)
}

func (s *streamWS) Context() context.Context {
	return NewContextWithServerTransportStream(s.ctx, s, s.method.grpcFullMethod)
}

func (s *streamWS) msgType() websocket.MessageType {
	if s.encoding == wsEncodingProto {
		return websocket.MessageBinary
	}
	return websocket.MessageText
}

func (s *streamWS) marshal(msg proto.Message) ([]byte, error) {
	if s.encoding == wsEncodingProto {
		return proto.Marshal(msg)
	}
	return protojson.Default.Marshal(msg)
}

func (s *streamWS) unmarshal(data []byte, msg proto.Message) error {
	if s.encoding == wsEncodingProto {
		return proto.Unmarshal(data, msg)
	}
	return protojson.Default.Unmarshal(data, msg)
}

func (s *streamWS) SendMsg(m any) error {
	if funk.IsNil(m) {
		return errors.New("stream ws send msg got nil")
	}

	reply, ok := m.(proto.Message)
	if !ok {
		return errors.New("stream ws send proto msg got unknown type message")
	}

	cur := reply.ProtoReflect()
	for _, fd := range getRspBodyDesc(s.pathOperation()) {
		cur = cur.Mutable(fd).Message()
	}
	msg := cur.Interface()

	b, err := s.marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal ws response")
	}

	if err = s.conn.Write(s.ctx, s.msgType(), b); err != nil {
		return errors.WrapCaller(err)
	}
	return nil
}

func (s *streamWS) RecvMsg(m any) error {
	if funk.IsNil(m) {
		return errors.New("stream ws recv msg got nil")
	}

	args, ok := m.(proto.Message)
	if !ok {
		return errors.New("stream ws recv proto msg got unknown type message")
	}

	_, data, err := s.conn.Read(s.ctx)
	if err != nil {
		// A normal websocket closure marks the end of the client stream.
		if status := websocket.CloseStatus(err); status != -1 {
			return io.EOF
		}
		// A cancelled/expired context also terminates the client stream.
		if s.ctx.Err() != nil {
			return io.EOF
		}
		return errors.WrapCaller(err)
	}

	cur := args.ProtoReflect()
	for _, fd := range getReqBodyDesc(s.pathOperation()) {
		cur = cur.Mutable(fd).Message()
	}
	msg := cur.Interface()

	if len(data) > 0 {
		if err = s.unmarshal(data, msg); err != nil {
			return errors.Wrapf(err, "failed to unmarshal ws request, msg=%#v", msg)
		}
	}

	// Merge path/query params into the first request message only.
	if !s.paramsMerged && len(s.params) > 0 {
		s.paramsMerged = true
		if err = gatewayutils.PopulateQueryParameters(args, s.params, gatewayutils.NewDoubleArray(nil)); err != nil {
			return errors.Wrapf(err, "failed to set ws query params, params=%v", s.params)
		}
	}

	return nil
}

// pathOperation is currently nil for websocket streams (no REST body rules), but
// kept as a hook so request/response body field descriptors can be wired later.
func (s *streamWS) pathOperation() *MatchOperation { return nil }
