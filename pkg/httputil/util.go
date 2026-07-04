package httputil

import (
	"log/slog"
	"strings"

	"dario.cat/mergo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/env"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/errors/errcode"
	"github.com/pubgo/funk/v2/proto/errorpb"
	"github.com/samber/lo"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/codes"

	"github.com/pubgo/lava/v2/pkg/encoding/protojson"
	"github.com/pubgo/lava/v2/pkg/fiberbuilder"
)

// DefaultBodyLimit is the default max HTTP request body size (4 MiB).
const DefaultBodyLimit = 4 << 20

const defaultHTTPPort = 8080

type Config struct {
	Http              *fiberbuilder.Config `yaml:"http"`
	EnablePrintRouter bool                 `yaml:"enable_print_router"`
	HttpPort          *int                 `yaml:"http_port"`
}

func DefaultCfg(config ...*Config) Config {
	cfg := Config{
		Http: &fiberbuilder.Config{
			ServerHeader:       "lava",
			EnableIPValidation: true,
			ETag:               true,
			ErrorHandler:       ErrHandler,
			BodyLimit:          DefaultBodyLimit,
		},
		EnablePrintRouter: true,
		HttpPort:          lo.ToPtr(defaultHTTPPort),
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
	return cfg
}

func IsWebsocket(h *fasthttp.RequestHeader) bool {
	if strings.Contains(strings.ToLower(string(h.Peek("Connection"))), "upgrade") &&
		strings.EqualFold(string(h.Peek("Upgrade")), "websocket") {
		return true
	}
	return false
}

func ErrHandler(ctx fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	var errPb *errorpb.ErrCode
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) && fiberErr != nil {
		errPb = &errorpb.ErrCode{
			Name:       "lava.error",
			StatusCode: errorpb.Code(errcode.Http2GrpcCode(int32(fiberErr.Code))),
			Code:       int32(fiberErr.Code),
			Message:    fiberErr.Message,
			Details: errcode.MustTagsToAny(errors.Tags{
				"path":     ctx.Route().Path,
				"version":  version.Version(),
				"instance": env.Get("HOSTNAME"),
			}),
		}
	} else {
		errPb = errcode.ParseError(err)
	}

	if errPb == nil || errPb.StatusCode == 0 {
		return nil
	}

	code := errcode.GrpcCodeToHTTP(codes.Code(errPb.StatusCode))
	ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	data, err := protojson.Default.Marshal(errPb)
	if err != nil {
		slog.Error("failed to marshal errorpb.ErrCode error", "err", err)
		return err
	}
	return ctx.Status(code).Send(data)
}

func Cors() fiber.Handler {
	cfg := cors.Config{
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
			fiber.MethodHead,
			fiber.MethodOptions,
		},
		MaxAge: 0,
	}

	// In dev/test, allow cross-origin without credentials. Production requires explicit config.
	if isNonProdRuntime() || debugEnabled() {
		cfg.AllowOriginsFunc = func(origin string) bool {
			return origin != ""
		}
	} else {
		cfg.AllowOriginsFunc = func(origin string) bool {
			return false
		}
	}

	return cors.New(cfg)
}

func isNonProdRuntime() bool {
	switch strings.ToLower(firstEnv("runenv", "env")) {
	case "", "dev", "test", "stage", "staging":
		return true
	default:
		return false
	}
}

func debugEnabled() bool {
	for _, key := range []string{"debug", "enable_debug"} {
		switch strings.ToLower(env.Get(key)) {
		case "true", "1":
			return true
		}
	}
	return false
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if val := env.Get(key); val != "" {
			return val
		}
	}
	return ""
}
