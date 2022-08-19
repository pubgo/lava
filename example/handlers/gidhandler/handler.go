package gidhandler

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/google/uuid"
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/errors"
	"github.com/pubgo/lava/example/pkg/proto/gidpb"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/service"
	"github.com/teris-io/shortid"
)

var _ service.Init = (*Id)(nil)
var _ service.Gateway = (*Id)(nil)

type Id struct {
	Cron   *scheduler.Scheduler
	Metric metric.Metric

	snowflake *snowflake.Snowflake
	bigflake  *bigflake.Bigflake
}

func (id *Id) Gateway() func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return gidpb.RegisterEchoServiceHandler
}

func New() gidpb.IdServer {
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
		snowflake: sf,
		bigflake:  bg,
	}
}

func (id *Id) Init() error {
	id.Cron.Every("test gid", time.Second*2, func(name string) {
		//id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Counter(name).Inc(1)
		//id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Gauge(name).Update(1)
		id.Metric.Tagged(metric.Tags{"module": "scheduler"}).Gauge(name).Update(1)
		fmt.Println("test cron every")
	})
	return nil
}

var err1 = errors.New("id.generate")

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