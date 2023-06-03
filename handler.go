package lava

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/opendoc/opendoc"
	"google.golang.org/grpc"
)

type GrpcHandler interface {
	Init()
	Internal() bool
	Version() string
	Middlewares() []Middleware
	ServiceDesc() *grpc.ServiceDesc
}

type HttpRouter interface {
	Init()
	Internal() bool
	Version() string
	Middlewares() []Middleware
	Router(app *fiber.App)
	Openapi(swag *opendoc.Swagger)
}

func WrapHandler[Req any, Rsp any](handle func(ctx context.Context, req *Req) (rsp *Rsp, err error)) func(ctx *fiber.Ctx) error {
	validate := validator.New()

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

		rsp, err := handle(ctx.UserContext(), &req)
		if err != nil {
			return err
		}

		return ctx.JSON(rsp)
	}
}
