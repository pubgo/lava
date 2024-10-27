package tasks

import (
	"fmt"
	"net"
	"net/http"
	"time"

	fiber "github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/lava/core/debug"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/internal/logutil"
	"github.com/pubgo/lava/lava"
	"google.golang.org/grpc/codes"
)

func New(srv lava.Service) *Server {
	assert.If(srv == nil, "service is nil")

	return &Server{srv: srv}
}

type Server struct {
	srv        lava.Service
	log        log.Logger
	lc         lifecycle.Getter
	httpServer *fiber.App
	conf       *Config
}

func (s *Server) Run() {
	defer s.stop()
	s.start()
	signal.Wait()
}

func (s *Server) DixInject(
	getLifecycle lifecycle.Getter,
	log log.Logger,
	conf []*Config,
) {
	s.lc = getLifecycle
	s.log = log.WithName("task-server")

	if len(conf) > 0 {
		s.conf = conf[0]
	} else {
		s.conf = &Config{HttpPort: generic.Ptr(running.HttpPort)}
	}

	s.httpServer = fiber.New(fiber.Config{
		EnableIPValidation: true,
		ErrorHandler: func(ctx fiber.Ctx, err error) error {
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

	s.httpServer.Use("/debug", debug.App())
}

func (s *Server) start() {
	defer recovery.Exit()

	logutil.OkOrFailed(s.log, "running before service starts", func() error {
		defer recovery.Exit()
		for _, run := range s.lc.GetBeforeStarts() {
			s.log.Info().Msgf("running %s", stack.CallerWithFunc(run.Handler))
			run.Handler()
		}
		return nil
	})

	httpLn := assert.Must1(net.Listen("tcp", fmt.Sprintf(":%d", generic.DePtr(s.conf.HttpPort))))

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

		async.GoSafe(func() error {
			s.srv.Start()
			return nil
		})
		return nil
	})

	logutil.OkOrFailed(s.log, "running after service starts", func() error {
		for _, run := range s.lc.GetAfterStarts() {
			logutil.LogOrErr(
				s.log,
				fmt.Sprintf("running %s", stack.CallerWithFunc(run.Handler)),
				func() error { run.Handler(); return nil },
			)
		}
		return nil
	})
}

func (s *Server) stop() {
	defer recovery.Exit()

	logutil.OkOrFailed(s.log, "running before service stops", func() error {
		for _, run := range s.lc.GetBeforeStops() {
			logutil.LogOrErr(
				s.log,
				fmt.Sprintf("running %s", stack.CallerWithFunc(run.Handler)),
				func() error { run.Handler(); return nil },
			)
		}
		return nil
	})

	logutil.LogOrErr(s.log, "[http-debug-server] Shutdown", func() error {
		return s.httpServer.ShutdownWithTimeout(time.Second * 5)
	})

	logutil.LogOrErr(s.log, "[task] Server Stop", func() error {
		s.srv.Stop()
		return nil
	})

	logutil.OkOrFailed(s.log, "running after service stops", func() error {
		for _, run := range s.lc.GetAfterStops() {
			logutil.LogOrErr(
				s.log,
				fmt.Sprintf("running %s", stack.CallerWithFunc(run.Handler)),
				func() error { run.Handler(); return nil },
			)
		}
		return nil
	})
}
