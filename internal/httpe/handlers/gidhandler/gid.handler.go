package gidhandler

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/funk/metric"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/opendoc/opendoc"
)

var _ service.HttpRouter = (*Id)(nil)

type Id struct {
	cron      *scheduler.Scheduler
	metric    metric.Metric
	snowflake *snowflake.Snowflake
	bigflake  *bigflake.Bigflake
}

func (id *Id) Router(app *fiber.App) {
	app.Get("/hello")
}

func (id *Id) Openapi(swag *opendoc.Swagger) {
	swag.ServiceOf("sss", func(srv *opendoc.Service) {

	})
}

func (id *Id) Middlewares() []service.Middleware {
	return nil
}

func New(cron *scheduler.Scheduler, metric metric.Metric) service.HttpRouter {
	id := rand.Intn(100)

	sf, err := snowflake.New(uint32(id))
	if err != nil {
		panic(err.Error())
	}
	bg, err := bigflake.New(uint64(id))
	if err != nil {
		panic(err.Error())
	}

	return &Id{
		cron:      cron,
		metric:    metric,
		snowflake: sf,
		bigflake:  bg,
	}
}

func (id *Id) Init() {
	id.cron.Every("test gid", time.Second*2, func(name string) {
		//id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Counter(name).Inc(1)
		//id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Gauge(name).Update(1)
		id.metric.Tagged(metric.Tags{"module": "scheduler"}).Counter(name).Inc(1)
		fmt.Println("test cron every")
	})
}
