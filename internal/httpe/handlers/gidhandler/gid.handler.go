package gidhandler

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/lava/internal/httpe/pkg/gidpb"
	"github.com/teris-io/shortid"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/funk/metric"
	"github.com/pubgo/lava/core/scheduler"
	service "github.com/pubgo/lava/lava"
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
	swag.ServiceOf("http", func(srv *opendoc.Service) {
		srv.GetOf(func(op *opendoc.Operation) {
			op.SetModel(new(gidpb.GenerateRequest), new(gidpb.GenerateResponse))
			op.SetPath()
		})
	})
}

func (id *Id) Generate(ctx context.Context, req *gidpb.GenerateRequest) (*gidpb.GenerateResponse, error) {
	var rsp = new(gidpb.GenerateResponse)
	var logs = log.Ctx(ctx)

	if len(req.Type) == 0 {
		req.Type = "uuid"
	}

	switch req.Type {
	case "uuid":
		rsp.Type = "uuid"
		rsp.Id = uuid.New().String()
	case "snowflake":
		id, err := id.snowflake.Mint()
		if err != nil {
			logs.Err(err).Msg("Failed to generate snowflake id")
			err = errors.Wrap(err, "Failed to generate snowflake id")
			return nil, errors.WrapCode(err, gidpb.ErrSrvErrCodeIDGenerateFailed)
		}
		rsp.Type = "snowflake"
		rsp.Id = fmt.Sprintf("%v", id)
	case "bigflake":
		id, err := id.bigflake.Mint()
		if err != nil {
			logs.Err(err).Msg("Failed to generate bigflake id")
			err = errors.Wrap(err, "failed to mint bigflake id")
			return nil, errors.WrapCode(err, gidpb.ErrSrvErrCodeIDGenerateFailed)
		}
		rsp.Type = "bigflake"
		rsp.Id = fmt.Sprintf("%v", id)
	case "shortid":
		id, err := shortid.Generate()
		if err != nil {
			logs.Err(err).Msg("Failed to generate shortid id")
			err = errors.Wrap(err, "failed to generate short id")
			return nil, errors.WrapCode(err, gidpb.ErrSrvErrCodeIDGenerateFailed)
		}
		rsp.Type = "shortid"
		rsp.Id = id
	default:
		return nil, errors.WrapCode(errors.New("unsupported id type"), gidpb.ErrSrvErrCodeIDGenerateFailed)
	}

	return rsp, nil
}

func (id *Id) Types(ctx context.Context, req *gidpb.TypesRequest) (*gidpb.TypesResponse, error) {
	var rsp = new(gidpb.TypesResponse)
	rsp.Types = []string{
		"uuid",
		"shortid",
		"snowflake",
		"bigflake",
	}
	return rsp, nil
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
