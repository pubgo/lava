package httprouter

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Handler[Req any, Rsp any] func(ctx *fiber.Ctx, req *Req) (rsp *Rsp, err error)

var validate = validator.New()

func WrapHandler[Req, Rsp any](handle func(ctx *fiber.Ctx, req *Req) (rsp *Rsp, err error)) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		var req Req

		if err := ctx.ParamsParser(&req); err != nil {
			return fmt.Errorf("failed to parse params, err:%w", err)
		}

		if err := ctx.QueryParser(&req); err != nil {
			return fmt.Errorf("failed to parse query, err:%w", err)
		}

		if err := ctx.ReqHeaderParser(&req); err != nil {
			return fmt.Errorf("failed to parse req header, err:%w", err)
		}

		switch ctx.Method() {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			if err := ctx.BodyParser(&req); err != nil {
				return fmt.Errorf("failed to parse body, err:%w", err)
			}
		}

		if err := validate.Struct(&req); err != nil {
			return fmt.Errorf("failed to validate request, err:%w", err)
		}

		rsp, err := handle(ctx, &req)
		if err != nil {
			return err
		}

		if rsp == nil {
			return ctx.JSON(make(map[string]any))
		}

		return ctx.JSON(rsp)
	}
}
