package fiber

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ grpc.ServerStream = (*wsStream)(nil)

func NewWsStream(ctx *fiber.Ctx, conn *Conn) grpc.ServerStream {
	return &wsStream{
		ctx:         ctx,
		conn:        conn,
		messageType: websocket.TextMessage,
		md:          metadata.MD{},
	}
}

type wsStream struct {
	ctx         *fiber.Ctx
	conn        *Conn
	messageType int
	md          metadata.MD
}

func (w *wsStream) SetHeader(md metadata.MD) error {
	for k, v := range md {
		w.md.Set(k, v...)
	}
	return nil
}

func (w *wsStream) SendHeader(md metadata.MD) error {
	return w.conn.WriteJSON(md)
}

func (w *wsStream) SetTrailer(md metadata.MD) {
	_ = w.conn.WriteJSON(md)
}

func (w *wsStream) Context() context.Context {
	return w.ctx.Context()
}

func (w *wsStream) SendMsg(m interface{}) error {
	if err := w.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		return err
	}

	return w.conn.WriteJSON(m)
}

func (w *wsStream) RecvMsg(m interface{}) error {
	mt, msg, err := w.conn.ReadMessage()
	if err != nil {
		if IsUnexpectedCloseError(err, CloseGoingAway, CloseAbnormalClosure) {
			log.Printf("error: %v", err)
			return nil
		}
		return err
	}

	w.messageType = mt
	if err := json.Unmarshal(msg, m); err != nil {
		return xerror.Wrap(err)
	}

	return nil
}
