package https

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/lava"
)

func init() {
	fiber.SetParserDecoder(fiber.ParserConfig{
		IgnoreUnknownKeys: true,
		ZeroEmpty:         true,
	})
}

func RegParser(parsers []fiber.ParserType) {
	fiber.SetParserDecoder(fiber.ParserConfig{
		IgnoreUnknownKeys: true,
		ZeroEmpty:         true,
		ParserType:        parsers,
	})
}

func handlerHttpMiddle(middlewares []lava.Middleware) func(fbCtx *fiber.Ctx) error {
	h := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		reqCtx := req.(*httpRequest)
		reqCtx.ctx.SetUserContext(ctx)
		return &httpResponse{ctx: reqCtx.ctx}, reqCtx.ctx.Next()
	}

	h = lava.Chain(middlewares...).Middleware(h)
	return func(ctx *fiber.Ctx) error {
		_, err := h(ctx.Context(), &httpRequest{ctx: ctx})
		return err
	}
}
