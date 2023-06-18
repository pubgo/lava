package https

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava"
)

// DefaultMaxBodyBytes is the maximum allowed size of a request body in bytes.
const DefaultMaxBodyBytes = 256 * 1024

var parserTypes []fiber.ParserType

func RegParserType(customType interface{}, converter func(string) reflect.Value) {
	parserTypes = append(parserTypes, fiber.ParserType{
		Customtype: customType,
		Converter:  converter,
	})
}

var validate = validator.New()

func Handler[Req any, Rsp any](hh func(ctx context.Context, req *Req) (rsp *Rsp, err error)) func(ctx *fiber.Ctx) error {
	// TODO check tag
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

		rsp, err := hh(ctx.Context(), &req)
		if err != nil {
			return err
		}

		return ctx.JSON(rsp)
	}
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
