package restEntry

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/types"
)

func (t *restEntry) handlerMiddle(middlewares []types.Middleware) func(fbCtx *fiber.Ctx) error {
	var handler = func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error {
		var reqCtx = req.(*httpRequest)

		for k, v := range reqCtx.Header() {
			for i := range v {
				reqCtx.ctx.Request().Header.Add(k, v[i])
			}
		}
		if err := reqCtx.ctx.Next(); err != nil {
			return err
		}

		reqCtx.ctx.SetUserContext(ctx)
		return rsp(&httpResponse{header: reqCtx.header, ctx: reqCtx.ctx})
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return func(fbCtx *fiber.Ctx) error {
		request := &httpRequest{
			ctx:    fbCtx,
			header: convertHeader(&fbCtx.Request().Header),
		}

		return handler(fbCtx.Context(), request, func(_ types.Response) error { return nil })
	}
}
