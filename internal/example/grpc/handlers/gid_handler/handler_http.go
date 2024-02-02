package gid_handler

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/internal/example/grpc/services/gid_client"
	"github.com/pubgo/lava/lava"
	"math/rand"
)

type Rsp struct {
	SS lava.HttpRouter
}

type IdHttp struct {
	cron      *scheduler.Scheduler
	metric    metrics.Metric
	snowflake *snowflake.Snowflake
	bigflake  *bigflake.Bigflake
	log       log.Logger
	service   *gid_client.Service
}

func (id *IdHttp) Annotation() []lava.Annotation {
	return nil
}

func (id *IdHttp) Router(router *lava.Router) {
	router.R.Get("/test123", func(ctx *fiber.Ctx) error {
		ctx.WriteString("hello world")
		return nil
	})
}

func (id *IdHttp) Middlewares() []lava.Middleware {
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

func NewHttp(cron *scheduler.Scheduler, metric metrics.Metric, log log.Logger, service *gid_client.Service) Rsp {
	id := rand.Intn(100)

	sf, err := snowflake.New(uint32(id))
	if err != nil {
		panic(err.Error())
	}
	bg, err := bigflake.New(uint64(id))
	if err != nil {
		panic(err.Error())
	}

	return Rsp{
		SS: &IdHttp{
			service:   service,
			cron:      cron,
			metric:    metric,
			snowflake: sf,
			bigflake:  bg,
			log:       log.WithName("gid"),
		},
	}
}
