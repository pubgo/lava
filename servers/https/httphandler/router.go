package httphandler

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type Handler[Req any, Rsp any] func(ctx fiber.Ctx, req *Req) (rsp *Rsp, err error)

var validate = validator.New()

func WrapHandler[Req, Rsp any](handler Handler[Req, Rsp]) func(ctx fiber.Ctx) error {
	return func(ctx fiber.Ctx) error {
		var req Req

		bind := ctx.Bind()
		if err := bind.URI(&req); err != nil {
			return fmt.Errorf("failed to parse params, params:%v err:%w", ctx.Route().Params, err)
		}

		if err := bind.Query(&req); err != nil {
			return fmt.Errorf("failed to parse query, query:%v err:%w", ctx.Queries(), err)
		}

		if err := bind.Header(&req); err != nil {
			return fmt.Errorf("failed to parse header, header:%q err:%w", ctx.GetReqHeaders(), err)
		}

		switch ctx.Method() {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			if err := bind.Body(&req); err != nil {
				return fmt.Errorf("failed to parse body, err:%w", err)
			}
		}

		if err := validate.Struct(&req); err != nil {
			return fmt.Errorf("failed to validate request, err:%w", err)
		}

		rsp, err := handler(ctx, &req)
		if err != nil {
			return err
		}

		if rsp == nil {
			return ctx.JSON(make(map[string]any))
		}

		return ctx.JSON(rsp)
	}
}
