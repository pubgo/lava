package fiber

import (
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ grpc.ServerStream = (*wsStream)(nil)

func NewWsStream(ctx *fiber.Ctx, conn *Conn) grpc.ServerStream {
	return &wsStream{ctx: ctx, conn: conn, messageType: websocket.TextMessage}
}

type wsStream struct {
	ctx         *fiber.Ctx
	conn        *Conn
	messageType int
}

func (w *wsStream) SetHeader(md metadata.MD) error {
	panic("implement me")
}

func (w *wsStream) SendHeader(md metadata.MD) error {
	panic("implement me")
}

func (w *wsStream) SetTrailer(md metadata.MD) {
	panic("implement me")
}

func (w *wsStream) Context() context.Context {
	return w.ctx.Context()
}

func (w *wsStream) SendMsg(m interface{}) error {
	return w.conn.WriteJSON(m)
}

func (w *wsStream) RecvMsg(m interface{}) error {
	mt, msg, err := w.conn.ReadMessage()
	if err != nil {
		return err
	}

	w.messageType = mt
	if err := json.Unmarshal(msg, m); err != nil {
		return xerror.Wrap(err)
	}

	return nil
}
