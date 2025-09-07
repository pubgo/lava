package httputil

import (
	"log/slog"
	"strings"
	
	"dario.cat/mergo"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pubgo/funk"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/proto/errorpb"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/v2/pkg/fiberbuilder"
	"github.com/samber/lo"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/codes"
)

type Config struct {
	Http              *fiberbuilder.Config `yaml:"http"`
	EnablePrintRouter bool                 `yaml:"enable_print_router"`
	BaseUrl           string               `yaml:"base_url"`
	HttpPort          *int                 `yaml:"http_port"`
}

func DefaultCfg(config ...*Config) Config {
	cfg := Config{
		Http: &fiberbuilder.Config{
			ServerHeader:       "lava",
			EnableIPValidation: true,
			ETag:               true,
			ErrorHandler:       ErrHandler,
			BodyLimit:          1024 * 1024 * 500,
		},
		EnablePrintRouter: true,
		BaseUrl:           version.Project(),
		HttpPort:          lo.ToPtr(running.HttpPort),
	}

	for _, t := range config {
		if t == nil {
			continue
		}

		err := mergo.Merge(&cfg, t, mergo.WithOverride, mergo.WithAppendSlice)
		if err != nil {
			slog.Error("failed to merge config to default config", "err", err, "source", t, "target", cfg)
			panic(err)
		}
	}

	cfg.Http.EnablePrintRoutes = cfg.EnablePrintRouter
	cfg.BaseUrl = funk.DoFunc(func() string {
		baseUrl := cfg.BaseUrl
		if baseUrl == "" {
			baseUrl = "/" + version.Project()
		}
		return "/" + strings.Trim(baseUrl, "/")
	})
	return cfg
}

func IsWebsocket(h *fasthttp.RequestHeader) bool {
	if strings.Contains(strings.ToLower(string(h.Peek("Connection"))), "upgrade") &&
		strings.EqualFold(string(h.Peek("Upgrade")), "websocket") {
		return true
	}
	return false
}

func ErrHandler(ctx *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	var errPb *errorpb.Error
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) && fiberErr != nil {
		errPb = &errorpb.Error{
			Code: &errorpb.ErrCode{
				Name:       "lava.error",
				StatusCode: errorpb.Code(errutil.Http2GrpcCode(int32(fiberErr.Code))),
				Code:       int32(fiberErr.Code),
				Message:    fiberErr.Message,
				Details: errors.MustTagsToAny(
					&errorpb.Tag{Key: "path", Value: ctx.Route().Path},
					&errorpb.Tag{Key: "version", Value: running.Version},
					&errorpb.Tag{Key: "instance", Value: running.InstanceID},
				),
			},
			Trace: &errorpb.ErrTrace{},
		}
	} else {
		errPb = errutil.ParseError(err)
	}

	if errPb == nil || errPb.Code.Code == 0 {
		return nil
	}

	errPb.Trace.Operation = ctx.Route().Path

	code := int(errPb.Code.Code)
	if errPb.Code.Code > 1000 {
		code = errutil.GrpcCodeToHTTP(codes.Code(errPb.Code.Code))
	}

	ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return ctx.Status(code).JSON(errPb)
}

func Cors() fiber.Handler {
	return cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
			fiber.MethodHead,
			fiber.MethodOptions,
		}, ","),
		//AllowHeaders:     "",
		AllowCredentials: true,
		//ExposeHeaders:    "",
		MaxAge: 0,
	})
}
