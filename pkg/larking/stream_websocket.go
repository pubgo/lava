package larking

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"net/http"
	"sync"
	"time"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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

const kindWebsocket = "WEBSOCKET"

type streamWS struct {
	ctx        context.Context
	conn       *websocket.Conn
	method     *method
	header     metadata.MD
	trailer    metadata.MD
	params     params
	recOnce    sync.Once
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
	sts := &serverTransportStream{s, s.method.name}
	return grpc.NewContextWithServerTransportStream(s.ctx, sts)
}

func (s *streamWS) SendMsg(v interface{}) error {
	reply := v.(proto.Message)
	//ctx := s.ctx

	cur := reply.ProtoReflect()
	for _, fd := range s.method.resp {
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

	if s.method.hasBody {
		cur := args.ProtoReflect()
		for _, fd := range s.method.body {
			cur = cur.Mutable(fd).Message()
		}

		msg := cur.Interface()
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			return err
		}

		// TODO: contentType check?
		// What marshalling options should we support?
		if err := protojson.Unmarshal(message, msg); err != nil {
			return errors.Wrap(err, "failed to unmarshal protobuf json message")
		}
	}

	s.recOnce.Do(func() {
		if err := s.params.set(args); err != nil {
			log.Err(err).Msg("failed to set params")
		}
	})

	return nil
}
