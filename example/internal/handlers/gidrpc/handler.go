package gidrpc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/errors"
	"github.com/pubgo/lava/example/gen/proto/gidpb"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/service"
	"github.com/teris-io/shortid"
)

var err1 = errors.New("id.generate")

var _ service.Init = (*Id)(nil)
var _ service.GatewayRegister = (*Id)(nil)
var _ service.GrpcRegister[gidpb.IdServer] = (*Id)(nil)
var _ service.MiddlewareInject = (*Id)(nil)

type Id struct {
	cron      *scheduler.Scheduler
	m         metric.Metric
	snowflake *snowflake.Snowflake
	bigflake  *bigflake.Bigflake
}

func (id *Id) Middlewares() []service.Middleware {
	return []service.Middleware{func(next service.HandlerFunc) service.HandlerFunc {
		return func(ctx context.Context, req service.Request, rsp service.Response) error {
			if req.Service() != gidpb.Id_ServiceDesc.ServiceName {
				return next(ctx, req, rsp)
			}

			return next(ctx, req, rsp)
		}
	}}
}

func (id *Id) GrpcRegister() service.RegisterServer[gidpb.IdServer] {
	return gidpb.RegisterIdServer
}

func (id *Id) GatewayRegister() service.GatewayHandler {
	return gidpb.RegisterIdHandler
}

func (id *Id) Init() {
	id.cron.Every("test gid", time.Second*2, func(name string) {
		//id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Counter(name).Inc(1)
		//id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Gauge(name).Update(1)
		id.m.Tagged(metric.Tags{"module": "scheduler"}).Gauge(name).Update(1)
		fmt.Println("test cron every")
	})
}

func (id *Id) Generate(ctx context.Context, req *gidpb.GenerateRequest) (*gidpb.GenerateResponse, error) {
	var rsp = new(gidpb.GenerateResponse)
	var log = logging.GetLog(ctx)

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
			log.Sugar().Errorf("Failed to generate snowflake id: %v", err)
			return nil, err1.Msg("id.generate", "failed to mint snowflake id").StatusBadRequest()
		}
		rsp.Type = "snowflake"
		rsp.Id = fmt.Sprintf("%v", id)
	case "bigflake":
		id, err := id.bigflake.Mint()
		if err != nil {
			log.Sugar().Errorf("Failed to generate bigflake id: %v", err)
			return nil, err1.Msg("id.generate", "failed to mint bigflake id").StatusBadRequest()
		}
		rsp.Type = "bigflake"
		rsp.Id = fmt.Sprintf("%v", id)
	case "shortid":
		id, err := shortid.Generate()
		if err != nil {
			log.Sugar().Errorf("Failed to generate shortid id: %v", err)
			return nil, err1.Msg("id.generate", "failed to generate short id").StatusBadRequest()
		}
		rsp.Type = "shortid"
		rsp.Id = id
	default:
		return nil, err1.Msg("id.generate", "unsupported id type").StatusBadRequest()
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
