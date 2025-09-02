package tasks

import (
	"context"
	"fmt"
	"github.com/pubgo/funk/async"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/vars"
	"github.com/rs/xid"
	"github.com/samber/lo"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/supervisor"
	"github.com/pubgo/lava/v2/internal/logutil"
	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/netutil"
)

type Params struct {
	Log  log.Logger
	Conf []*Config
}

func New(params Params) supervisor.Service {
	s := &Server{}
	s.init(params.Log, params.Conf)
	return supervisor.NewService(s.String(), s.Serve)
}

type Server struct {
	log        log.Logger
	httpServer *fiber.App
	conf       *Config
}

func (s *Server) String() string {
	return "tasks"
}

func (s *Server) Serve(ctx context.Context) error {
	defer func() {
		logutil.LogOrErr(s.log, "[http-server] Shutdown", func() error {
			err := s.httpServer.ShutdownWithContext(ctx)
			if netutil.IsErrServerClosed(err) {
				return nil
			}
			return err
		})
	}()

	addr := fmt.Sprintf(":%d", generic.FromPtr(s.conf.HttpPort))
	s.log.Info().Msg("[http-server] Server Starting")
	async.GoDelay(func() error {
		err := s.httpServer.Listen(addr)
		if netutil.IsErrServerClosed(err) {
			return nil
		}
		return err
	})

	<-ctx.Done()
	return nil
}

func (s *Server) init(log log.Logger, conf []*Config) {
	s.log = log.WithName(s.String())
	s.conf = lo.ToPtr(httputil.DefaultCfg(conf...))

	vars.RegisterValue(s.String()+"_config_"+xid.New().String(), s.conf)

	cfg := s.conf.Http.Build().Must()
	s.httpServer = fiber.New(cfg)
	s.httpServer.Use(httputil.Cors())
	s.httpServer.Mount("/debug", debug.App())
}
