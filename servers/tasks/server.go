package tasks

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/stack"
	"google.golang.org/grpc/codes"

	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/supervisor"
	"github.com/pubgo/lava/internal/logutil"
)

type Params struct {
	Services     []supervisor.Service
	GetLifecycle lifecycle.Getter
	Log          log.Logger
	Conf         []*Config
}

func New(params Params) *supervisor.Manager {
	assert.If(len(params.Services) == 0, "services is nil")

	s := &Server{}
	s.init(
		params.GetLifecycle,
		params.Log,
		params.Conf,
	)

	manager := supervisor.Default()
	assert.Exit(manager.Add(supervisor.NewService("task", s.Serve)))
	for _, srv := range params.Services {
		assert.Exit(manager.Add(srv))
	}
	return manager
}

type Server struct {
	log        log.Logger
	lc         lifecycle.Getter
	httpServer *fiber.App
	conf       *Config
}

func (s *Server) String() string {
	return "task"
}

func (s *Server) Serve(ctx context.Context) error {
	defer s.stop(ctx)
	s.start(ctx)
	<-ctx.Done()
	return nil
}

func (s *Server) init(
	getLifecycle lifecycle.Getter,
	log log.Logger,
	conf []*Config,
) {
	s.lc = getLifecycle
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

func (s *Server) start(ctx context.Context) {
	defer recovery.Exit()

	logutil.OkOrFailed(s.log, "running before service start", func() error {
		defer recovery.Exit()
		for _, run := range s.lc.GetBeforeStarts() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Exec))
			assert.Exit(run.Exec(ctx))
		}
		return nil
	})

	httpLn := assert.Exit1(net.Listen("tcp", fmt.Sprintf(":%d", generic.FromPtr(s.conf.HttpPort))))

	logutil.OkOrFailed(s.log, "service starts", func() error {
		async.GoDelay(func() error {
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
			return nil
		})

		return nil
	})

	logutil.OkOrFailed(s.log, "running after service starts", func() error {
		for _, run := range s.lc.GetAfterStarts() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Exec))
			assert.Exit(run.Exec(ctx))
		}
		return nil
	})
}

func (s *Server) stop(ctx context.Context) {
	defer recovery.DebugPrint()

	logutil.OkOrFailed(s.log, "running before service stops", func() error {
		for _, run := range s.lc.GetBeforeStops() {
			logutil.LogOrErr(
				s.log,
				fmt.Sprintf("running %s", stack.CallerWithFunc(run.Exec)),
				func() error { return run.Exec(ctx) },
			)
		}
		return nil
	})

	logutil.LogOrErr(s.log, "[http-debug-server] Shutdown", func() error {
		return s.httpServer.ShutdownWithTimeout(time.Second * 5)
	})

	logutil.OkOrFailed(s.log, "running after service stops", func() error {
		for _, run := range s.lc.GetAfterStops() {
			logutil.LogOrErr(
				s.log,
				fmt.Sprintf("running %s", stack.CallerWithFunc(run.Exec)),
				func() error { return run.Exec(ctx) },
			)
		}
		return nil
	})
}
