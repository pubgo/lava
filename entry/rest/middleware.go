package rest

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lug/types"
)

func (t *restEntry) middleWrapper(ctx context.Context, req types.Request, rsp func(response types.Response) error) error {
	var reqCtx = req.(*httpRequest).ctx
	reqCtx.SetUserContext(ctx)

	if err := reqCtx.Next(); err != nil {
		return err
	}

	return rsp(&httpResponse{ctx: reqCtx})
}

func (t *restEntry) handlerLugMiddle(fbCtx *fiber.Ctx) error {
	t.middleOnce.Do(func() {
		var middlewares = t.Options().Middlewares
		for i := len(middlewares) - 1; i >= 0; i-- {
			t.handler = middlewares[i](t.middleWrapper)
		}
	})

	request := &httpRequest{ctx: fbCtx}
	return t.handler(fbCtx.Context(), request, func(_ types.Response) error { return nil })
}
