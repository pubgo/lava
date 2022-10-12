package https

import (
	"context"

	"github.com/go-playground/validator/v10"
	_ "github.com/go-playground/validator/v10"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/service"
)

func Wrap(hh handler) func(ctx *fiber.Ctx) error {
	var validate = validator.New()
	return func(ctx *fiber.Ctx) error {
		ctx.ParamsParser()
		ctx.QueryParser()
		ctx.BodyParser()
		ctx.ReqHeaderParser()

		var rsp, err = hh(ctx.Context(), nil)
		if err != nil {
			return err
		}

		err := validate.Struct(s)
		return ctx.JSON(rsp)
	}
}

type handler func(ctx context.Context, req interface{}) (rsp interface{}, err error)

func handlerHttpMiddle(middlewares []service.Middleware) func(fbCtx *fiber.Ctx) error {
	var h = func(ctx context.Context, req service.Request, rsp service.Response) error {
		var reqCtx = req.(*httpRequest)
		reqCtx.ctx.SetUserContext(ctx)
		return reqCtx.ctx.Next()
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return func(fbCtx *fiber.Ctx) error {
		return h(fbCtx.Context(), &httpRequest{ctx: fbCtx}, &httpResponse{ctx: fbCtx})
	}
}
