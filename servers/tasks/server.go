package tasks

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/core/supervisor"
	"github.com/pubgo/lava/internal/logutil"
	"google.golang.org/grpc/codes"
)

type Params struct {
	Log  log.Logger
	Conf []*Config
}

func New(params Params) supervisor.Service {
	s := &Server{}
	s.init(params.Log, params.Conf)
	return supervisor.NewService("tasks", s.Serve)
}

type Server struct {
	log        log.Logger
	httpServer *fiber.App
	conf       *Config
}

func (s *Server) Serve(ctx context.Context) error {
	defer func() {
		logutil.LogOrErr(s.log, "[http-debug-server] Shutdown", func() error {
			return s.httpServer.ShutdownWithTimeout(time.Second * 5)
		})
	}()

	httpLn := assert.Exit1(net.Listen("tcp", fmt.Sprintf(":%d", generic.FromPtr(s.conf.HttpPort))))
	s.log.Info().Msg("[http-debug-server] Server Starting")
	logutil.LogOrErr(s.log, "[http-debug-server] Server Stop", func() error {
		defer recovery.Exit()
		if err := s.httpServer.Listener(httpLn); err != nil &&
			!errors.Is(err, http.ErrServerClosed) &&
			!errors.Is(err, net.ErrClosed) {
			return err
		}
		return nil
	})

	<-ctx.Done()
	return nil
}

func (s *Server) init(log log.Logger, conf []*Config) {
	s.log = log.WithName("tasks")

	if len(conf) > 0 {
		s.conf = conf[0]
	} else {
		s.conf = &Config{HttpPort: generic.Ptr(running.HttpPort)}
	}

	s.httpServer = fiber.New(fiber.Config{
		EnableIPValidation: true,
		ETag:               true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			if err == nil {
				return nil
			}

			errPb := errutil.ParseError(err)
			if errPb == nil || errPb.Code.Code == 0 {
				return nil
			}
			errPb.Trace.Operation = ctx.Route().Path
			code := errutil.GrpcCodeToHTTP(codes.Code(errPb.Code.Code))
			ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			return ctx.Status(code).JSON(errPb)
		},
	})

	s.httpServer.Mount("/debug", debug.App())
}
