package https

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/binder"

	"github.com/pubgo/lava/v2/lava"
)

func init() {
	binder.SetParserDecoder(binder.ParserConfig{
		IgnoreUnknownKeys: true,
		ZeroEmpty:         true,
	})
}

func RegParser(parsers []binder.ParserType) {
	binder.SetParserDecoder(binder.ParserConfig{
		IgnoreUnknownKeys: true,
		ZeroEmpty:         true,
		ParserType:        parsers,
	})
}

func handlerHttpMiddle(middlewares []lava.Middleware) func(fbCtx fiber.Ctx) error {
	h := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		reqCtx := req.(*httpRequest).ctx
		return &httpResponse{ctx: reqCtx}, reqCtx.Next()
	}

	h = lava.Chain(middlewares...).Middleware(h)
	return func(ctx fiber.Ctx) error {
		_, err := h(ctx, &httpRequest{ctx: ctx})
		return err
	}
}
