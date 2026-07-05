package serverhttp

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"github.com/pubgo/lava/v2/pkg/lava"
)

// HandlerMiddleware wraps lava middleware as a Fiber handler that adapts
// fiber.Ctx to lava.Request/Response (shared by gatewayserver and https).
func HandlerMiddleware(middlewares []lava.Middleware) func(fiber.Ctx) error {
	h := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		reqCtx := req.(*Request).Ctx
		return NewResponse(reqCtx), reqCtx.Next()
	}

	h = lava.Chain(middlewares...).Middleware(h)
	return func(ctx fiber.Ctx) error {
		_, err := h(ctx, NewRequest(ctx))
		return err
	}
}
