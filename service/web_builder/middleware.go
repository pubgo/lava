package service_builder

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/middleware"
)

func (t *webImpl) handlerHttpMiddle(middlewares []middleware.Middleware) func(fbCtx *fiber.Ctx) error {
	var handler = func(ctx context.Context, req middleware.Request, rsp middleware.Response) error {
		var reqCtx = req.(*httpRequest)
		reqCtx.ctx.SetUserContext(ctx)
		return reqCtx.ctx.Next()
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return func(fbCtx *fiber.Ctx) error {
		return handler(fbCtx.Context(), &httpRequest{ctx: fbCtx}, &httpResponse{ctx: fbCtx})
	}
}
