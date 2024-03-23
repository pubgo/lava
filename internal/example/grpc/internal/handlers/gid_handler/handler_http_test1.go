package gid_handler

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/gofiber/fiber/v2"
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/internal/example/grpc/internal/services/gid_client"
	"github.com/pubgo/lava/lava"
)

type IdHttp111 struct {
	cron      *scheduler.Scheduler
	metric    metrics.Metric
	snowflake *snowflake.Snowflake
	bigflake  *bigflake.Bigflake
	log       log.Logger
	service   *gid_client.Service
}

func (id *IdHttp111) Prefix() string {
	return "/test1"
}

func (id *IdHttp111) Annotation() []lava.Annotation {
	return nil
}

func (id *IdHttp111) Router(router fiber.Router) {
	router.Get("/test123111", func(ctx *fiber.Ctx) error {
		ctx.WriteString("hello world")
		return nil
	})
}

func (id *IdHttp111) Middlewares() []lava.Middleware {
	return lava.Middlewares{
		lava.MiddlewareWrap{
			Next: func(next lava.HandlerFunc) lava.HandlerFunc {
				return func(ctx context.Context, req lava.Request) (lava.Response, error) {
					id.log.Info().Msgf("middleware %s", req.Endpoint())
					fmt.Println(req.Header().String())
					return next(ctx, req)
				}
			},
			Name: "header",
		},
	}
}

func NewHttp111(cron *scheduler.Scheduler, metric metrics.Metric, log log.Logger, service *gid_client.Service) lava.HttpRouter {
	id := rand.Intn(100)

	sf, err := snowflake.New(uint32(id))
	if err != nil {
		panic(err.Error())
	}
	bg, err := bigflake.New(uint64(id))
	if err != nil {
		panic(err.Error())
	}

	return &IdHttp111{
		service:   service,
		cron:      cron,
		metric:    metric,
		snowflake: sf,
		bigflake:  bg,
		log:       log.WithName("gid"),
	}
}
