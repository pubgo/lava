package gidhandler

import (
	"context"
	"fmt"
	"github.com/pubgo/lava/internal/example/grpc/services/gidclient"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/teris-io/shortid"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/internal/example/grpc/pkg/proto/gidpb"
)

var _ lava.GrpcRouter = (*Id)(nil)
var _ lava.GrpcGatewayRouter = (*Id)(nil)

type Id struct {
	cron      *scheduler.Scheduler
	metric    metric.Metric
	snowflake *snowflake.Snowflake
	bigflake  *bigflake.Bigflake
	log       log.Logger
	service   *gidclient.Service
}

func (id *Id) RegisterGateway(ctx context.Context, mux *runtime.ServeMux, conn grpc.ClientConnInterface) error {
	return gidpb.RegisterIdHandlerClient(ctx, mux, gidpb.NewIdClient(conn))
}

func (id *Id) TypeStream(request *gidpb.TypesRequest, server gidpb.Id_TypeStreamServer) error {
	for i := 0; i < 5; i++ {
		rsp := new(gidpb.TypesResponse)
		rsp.Types = []string{
			"uuid",
			"shortid",
			"snowflake",
			"bigflake",
		}
		_ = server.Send(rsp)
	}
	return nil
}

func (id *Id) Middlewares() []lava.Middleware {
	return nil
}

func (id *Id) ServiceDesc() *grpc.ServiceDesc {
	return &gidpb.Id_ServiceDesc
}

func New(cron *scheduler.Scheduler, metric metric.Metric, log log.Logger, service *gidclient.Service) lava.GrpcRouter {
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
		service:   service,
		cron:      cron,
		metric:    metric,
		snowflake: sf,
		bigflake:  bg,
		log:       log.WithName("gid"),
	}
}

func (id *Id) Init() {
	fmt.Println("test cron every")
	id.cron.Every("test_gid", time.Second*2, func(ctx context.Context, name string) error {
		id.metric.Tagged(metric.Tags{"module": "scheduler"}).Counter(name).Inc(1)
		fmt.Println("test cron every")

		rsp, err := id.service.Types(ctx, &gidpb.TypesRequest{})
		if err != nil {
			return err
		}
		id.log.Info(ctx).Any("data", rsp.Types).Msg("Types")
		return nil
	})
}

func (id *Id) Generate(ctx context.Context, req *gidpb.GenerateRequest) (*gidpb.GenerateResponse, error) {
	rsp := new(gidpb.GenerateResponse)
	var typ = req.GetType().String()
	if len(typ) == 0 {
		typ = "uuid"
	}

	switch typ {
	case "uuid":
		rsp.Type = "uuid"
		rsp.Id = uuid.New().String()
	case "snowflake":
		da, err := id.snowflake.Mint()
		if err != nil {
			id.log.Err(err).Msg("Failed to generate snowflake id")
			err = errors.Wrap(err, "Failed to generate snowflake id")
			return nil, errors.WrapCode(err, gidpb.ErrCodeIDGenerateFailed)
		}
		rsp.Type = "snowflake"
		rsp.Id = fmt.Sprintf("%v", da)
	case "bigflake":
		da, err := id.bigflake.Mint()
		if err != nil {
			id.log.Err(err).Msg("Failed to generate bigflake id")
			err = errors.Wrap(err, "failed to mint bigflake id")
			return nil, errors.WrapCode(err, gidpb.ErrCodeIDGenerateFailed)
		}
		rsp.Type = "bigflake"
		rsp.Id = fmt.Sprintf("%v", da)
	case "shortid":
		da, err := shortid.Generate()
		if err != nil {
			id.log.Err(err).Msg("Failed to generate shortid id")
			err = errors.Wrap(err, "failed to generate short id")
			return nil, errors.WrapCode(err, gidpb.ErrCodeIDGenerateFailed)
		}
		rsp.Type = "shortid"
		rsp.Id = da
	default:
		return nil, errors.WrapCode(errors.New("unsupported id type"), gidpb.ErrCodeIDGenerateFailed)
	}

	return rsp, nil
}

func (id *Id) Types(ctx context.Context, req *gidpb.TypesRequest) (*gidpb.TypesResponse, error) {
	rsp := new(gidpb.TypesResponse)
	rsp.Types = []string{
		"uuid",
		"shortid",
		"snowflake",
		"bigflake",
	}
	return rsp, nil
}
