package gidhandler

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/opendoc/opendoc"
	"github.com/teris-io/shortid"
)

var _ lava.HttpRouter = (*Id)(nil)

type Id struct {
	log       log.Logger
	cron      *scheduler.Scheduler
	metric    metrics.Metric
	snowflake *snowflake.Snowflake
	bigflake  *bigflake.Bigflake
}

func (t *Id) Version() string {
	return "v1"
}

func (t *Id) Router(app fiber.Router) {
	app.Post("/id/generate", lava.WrapHandler(t.Generate))
	app.Get("/id/types", lava.WrapHandler(t.Types))
}

func (t *Id) Openapi(swag *opendoc.Swagger) {
	swag.ServiceOf("http", func(srv *opendoc.Service) {
		srv.GetOf(func(op *opendoc.Operation) {
			op.SetModel(new(GenerateRequest), new(GenerateResponse))
			op.SetPath("id.generate", "/v1/id/generate")
		})

		srv.GetOf(func(op *opendoc.Operation) {
			op.SetModel(new(TypesRequest), new(TypesResponse))
			op.SetPath("id.types", "/v1/id/types")
		})
	})
}

func (t *Id) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	rsp := new(GenerateResponse)
	if len(req.Type) == 0 {
		req.Type = "uuid"
	}

	switch req.Type {
	case "uuid":
		rsp.Type = "uuid"
		rsp.Id = uuid.New().String()
	case "snowflake":
		id, err := t.snowflake.Mint()
		if err != nil {
			t.log.Err(err, ctx).Msg("Failed to generate snowflake id")
			err = errors.Wrap(err, "Failed to generate snowflake id")
			return nil, errors.WrapCode(err, ErrSrvErrCodeIDGenerateFailed)
		}
		rsp.Type = "snowflake"
		rsp.Id = fmt.Sprintf("%v", id)
	case "bigflake":
		id, err := t.bigflake.Mint()
		if err != nil {
			t.log.Err(err, ctx).Msg("Failed to generate bigflake id")
			err = errors.Wrap(err, "failed to mint bigflake id")
			return nil, errors.WrapCode(err, ErrSrvErrCodeIDGenerateFailed)
		}
		rsp.Type = "bigflake"
		rsp.Id = fmt.Sprintf("%v", id)
	case "shortid":
		id, err := shortid.Generate()
		if err != nil {
			t.log.Err(err, ctx).Msg("Failed to generate shortid id")
			err = errors.Wrap(err, "failed to generate short id")
			return nil, errors.WrapCode(err, ErrSrvErrCodeIDGenerateFailed)
		}
		rsp.Type = "shortid"
		rsp.Id = id
	default:
		return nil, errors.WrapCode(errors.New("unsupported id type"), ErrSrvErrCodeIDGenerateFailed)
	}

	return rsp, nil
}

func (t *Id) Types(ctx context.Context, req *TypesRequest) (*TypesResponse, error) {
	rsp := new(TypesResponse)
	rsp.Types = []string{
		"uuid",
		"shortid",
		"snowflake",
		"bigflake",
	}
	return rsp, nil
}

func (t *Id) Middlewares() []lava.Middleware {
	return nil
}

func New(cron *scheduler.Scheduler, metric metrics.Metric) lava.HttpRouter {
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

func (t *Id) Init() {
	t.cron.Every("test gid", time.Second*2, func(ctx context.Context, name string) error {
		// id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Counter(name).Inc(1)
		// id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Gauge(name).Update(1)
		t.metric.Tagged(metrics.Tags{"module": "scheduler"}).Counter(name).Inc(1)
		fmt.Println("test cron every")
		return nil
	})
}
