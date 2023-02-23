package service

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
	Middlewares() []Middleware
	ServiceDesc() *grpc.ServiceDesc
	Gateway(opts []interface{}) TwirpServer
}

type HttpRouter interface {
	Init()
	Middlewares() []Middleware
	Router(app *fiber.App)
	Openapi(swag *opendoc.Swagger)
}

type TwirpServer interface {
	http.Handler

	// ServiceDescriptor returns gzipped bytes describing the .proto file that
	// this service was generated from. Once unzipped, the bytes can be
	// unmarshalled as a
	// google.golang.org/protobuf/types/descriptorpb.FileDescriptorProto.
	//
	// The returned integer is the index of this particular service within that
	// FileDescriptorProto's 'Service' slice of ServiceDescriptorProtos. This is a
	// low-level field, expected to be used for reflection.
	ServiceDescriptor() ([]byte, int)

	// ProtocGenTwirpVersion is the semantic version string of the version of
	// twirp used to generate this file.
	ProtocGenTwirpVersion() string

	// PathPrefix returns the HTTP URL path prefix for all methods handled by this
	// service. This can be used with an HTTP mux to route Twirp requests.
	// The path prefix is in the form: "/<prefix>/<package>.<Service>/"
	// that is, everything in a Twirp route except for the <Method> at the end.
	PathPrefix() string
}

func WrapHandler[Req any, Rsp any](handle func(ctx context.Context, req *Req) (rsp *Rsp, err error)) func(ctx *fiber.Ctx) error {
	var validate = validator.New()

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

		var rsp, err = handle(ctx.Context(), &req)
		if err != nil {
			return err
		}

		return ctx.JSON(rsp)
	}
}
