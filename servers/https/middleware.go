package https

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/errors/errutil"
	"google.golang.org/grpc/codes"

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

func init() {
	//http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
	fiber.DefaultErrorHandler = func(c *fiber.Ctx, err error) error {
		if err == nil {
			return nil
		}

		code := fiber.StatusBadRequest
		var errPb = errutil.ParseError(err)
		if errPb == nil || errPb.Code == 0 {
			return nil
		}

		code = errutil.GrpcCodeToHTTP(codes.Code(errPb.Code))
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return c.Status(code).JSON(errPb)
	}

	fiber.SetParserDecoder(fiber.ParserConfig{
		IgnoreUnknownKeys: true,
		ZeroEmpty:         true,
		ParserType:        parserTypes,
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

		var rsp, err = hh(ctx.Context(), &req)
		if err != nil {
			return err
		}

		return ctx.JSON(rsp)
	}
}

func handlerHttpMiddle(middlewares []lava.Middleware) func(fbCtx *fiber.Ctx) error {
	var h = func(ctx context.Context, req lava.Request) (lava.Response, error) {
		var reqCtx = req.(*httpRequest)
		reqCtx.ctx.SetUserContext(ctx)
		return &httpResponse{ctx: reqCtx.ctx}, reqCtx.ctx.Next()
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return func(ctx *fiber.Ctx) error {
		_, err := h(ctx.Context(), &httpRequest{ctx: ctx})
		return err
	}
}
