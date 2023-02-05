package https

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/errors"
	"github.com/pubgo/lava/service"
	"google.golang.org/grpc/codes"
)

// DefaultMaxBodyBytes is the maximum allowed size of a request body in bytes.
const DefaultMaxBodyBytes = 256 * 1024

// Fields tags used by tonic.
const (
	QueryTag      = "query"
	PathTag       = "path"
	HeaderTag     = "header"
	EnumTag       = "enum"
	RequiredTag   = "required"
	DefaultTag    = "default"
	ValidationTag = "validate"
	ExplodeTag    = "explode"
)

const (
	DEFAULT     = "default"
	BINDING     = "binding"
	DESCRIPTION = "description"
	QUERY       = "query"
	FORM        = "form"
	URI         = "uri"
	HEADER      = "header"
	COOKIE      = "cookie"
)

// ParamIn defines parameter location.
type ParamIn string

const (
	// ParamInPath indicates path parameters, such as `/users/{id}`.
	ParamInPath = ParamIn("path")

	// ParamInQuery indicates query parameters, such as `/users?page=10`.
	ParamInQuery = ParamIn("query")

	// ParamInBody indicates body value, such as `{"id": 10}`.
	ParamInBody = ParamIn("body")

	// ParamInFormData indicates body form parameters.
	ParamInFormData = ParamIn("formData")

	// ParamInCookie indicates cookie parameters, which are passed ParamIn the `Cookie` header,
	// such as `Cookie: debug=0; gdpr=2`.
	ParamInCookie = ParamIn("cookie")

	// ParamInHeader indicates header parameters, such as `X-Header: value`.
	ParamInHeader = ParamIn("header")
)

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
		var errPb = errors.FromError(err)
		if errPb == nil || errPb.Code == 0 {
			return nil
		}

		code = errors.GrpcCodeToHTTP(codes.Code(errPb.Code))
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

func HandlerGet[Req any, Rsp any](app fiber.Router, prefix string, hh func(ctx context.Context, req *Req) (rsp *Rsp, err error), middlewares ...service.Middleware) {
	app.Get(prefix, handlerHttpMiddle(middlewares), Handler(hh))
}

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

func handlerHttpMiddle(middlewares []service.Middleware) func(fbCtx *fiber.Ctx) error {
	var h = func(ctx context.Context, req service.Request, rsp service.Response) error {
		var reqCtx = req.(*httpRequest)
		reqCtx.ctx.SetUserContext(ctx)
		return reqCtx.ctx.Next()
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return func(ctx *fiber.Ctx) error {
		return h(ctx.Context(), &httpRequest{ctx: ctx}, &httpResponse{ctx: ctx})
	}
}
