package https

import (
	"context"
	"reflect"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/lava"
)

var parserTypes []fiber.ParserType

func RegParser(customType interface{}, converter func(string) reflect.Value) {
	parserTypes = append(parserTypes, fiber.ParserType{
		Customtype: customType,
		Converter:  converter,
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
