package https

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/service"
)

var validate = validator.New()

func Wrap[Req any, Rsp any](hh func(ctx context.Context, req *Req) (rsp *Rsp, err error)) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var req Req

		if err := ctx.ParamsParser(&req); err != nil {
			return fmt.Errorf("failed to parse params, err:%w", err)
		}

		if err := ctx.QueryParser(&req); err != nil {
			return fmt.Errorf("failed to parse query, err:%w", err)
		}

		switch ctx.Method() {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			if err := ctx.BodyParser(&req); err != nil {
				return fmt.Errorf("failed to parse body, err:%w", err)
			}
		}

		if err := ctx.ReqHeaderParser(&req); err != nil {
			return fmt.Errorf("failed to parse req header, err:%w", err)
		}

		if err := validate.Struct(&req); err != nil {
			return fmt.Errorf("failed to validate request, err:%w", err)
		}

		var rsp, err = hh(ctx.Context(), &req)
		if err != nil {
			return err
		}

		return ctx.JSON(rsp)
	}
}

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
