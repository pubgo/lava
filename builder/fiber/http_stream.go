package fiber

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ grpc.ServerStream = (*httpStream)(nil)

func NewHttpStream(ctx *fiber.Ctx) grpc.ServerStream {
	return &wsStream{ctx: ctx}
}

type httpStream struct {
	ctx *fiber.Ctx
}

func (w *httpStream) SetHeader(md metadata.MD) error {
	for k, v := range md {
		for i := range v {
			w.ctx.Response().Header.Add(k, v[i])
		}
	}
	return nil
}

func (w *httpStream) SendHeader(md metadata.MD) error {
	return w.ctx.JSON(md)
}

func (w *httpStream) SetTrailer(md metadata.MD) {
	_ = w.ctx.JSON(md)
}

func (w *httpStream) Context() context.Context {
	return w.ctx.Context()
}

func (w *httpStream) SendMsg(m interface{}) error {
	return w.ctx.JSON(m)
}

func (w *httpStream) RecvMsg(m interface{}) error {
	return w.ctx.BodyParser(&m)
}
