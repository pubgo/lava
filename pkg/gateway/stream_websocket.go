package gateway

import (
	"context"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"net/url"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 10
)

type streamWS struct {
	ctx        *fiber.Ctx
	conn       *websocket.Conn
	pathRule   *httpPathRule
	header     metadata.MD
	trailer    metadata.MD
	params     url.Values
	sentHeader bool
}

func (s *streamWS) SetHeader(md metadata.MD) error {
	if !s.sentHeader {
		s.header = metadata.Join(s.header, md)
	}
	return nil
}

func (s *streamWS) SendHeader(md metadata.MD) error {
	if s.sentHeader {
		return nil // already sent?
	}
	// TODO: headers?
	s.sentHeader = true
	return nil
}

func (s *streamWS) SetTrailer(md metadata.MD) {
	s.sentHeader = true
	s.trailer = metadata.Join(s.trailer, md)
}

func (s *streamWS) Context() context.Context {
	//metadata.NewIncomingContext()
	sts := &serverTransportStream{ServerStream: s, method: s.pathRule.grpcMethodName}
	return grpc.NewContextWithServerTransportStream(s.ctx.Context(), sts)
}

func (s *streamWS) SendMsg(v interface{}) error {
	reply := v.(proto.Message)

	cur := reply.ProtoReflect()
	for _, fd := range s.pathRule.rspBody {
		cur = cur.Mutable(fd).Message()
	}
	msg := cur.Interface()

	// TODO: contentType check?
	b, err := protojson.Marshal(msg)
	if err != nil {
		return err
	}
	return s.conn.WriteMessage(websocket.TextMessage, b)
}

func (s *streamWS) RecvMsg(m interface{}) error {
	args := m.(proto.Message)

	if len(s.params) > 0 {
		if err := PopulateQueryParameters(args, s.params, utilities.NewDoubleArray(nil)); err != nil {
			log.Err(err).Msg("failed to set params")
		}
	}

	if s.pathRule.hasReqBody {
		cur := args.ProtoReflect()
		for _, fd := range s.pathRule.reqBody {
			cur = cur.Mutable(fd).Message()
		}

		msg := cur.Interface()
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			return err
		}

		if err := protojson.Unmarshal(message, msg); err != nil {
			return errors.Wrap(err, "failed to unmarshal protobuf json message")
		}
	}

	return nil
}
