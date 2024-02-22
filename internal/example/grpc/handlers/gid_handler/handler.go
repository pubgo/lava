package gid_handler

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/clients/resty"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/lava"
	"github.com/teris-io/shortid"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"math/rand"
	"net/http"
	"time"

	"github.com/pubgo/lava/internal/example/grpc/pkg/proto/gidpb"
	"github.com/pubgo/lava/internal/example/grpc/services/gid_client"
)

var typesReq = &resty.RequestConfig{
	Method: http.MethodGet,
	Path:   "/v1/id/types",
}

var _ lava.GrpcRouter = (*Id)(nil)
var _ gidpb.IdServer = (*Id)(nil)

type Id struct {
	cron      *scheduler.Scheduler
	metric    metrics.Metric
	snowflake *snowflake.Snowflake
	bigflake  *bigflake.Bigflake
	log       log.Logger
	service   *gid_client.Service
}

func (id *Id) Chat1(server gidpb.Id_Chat1Server) error {
	for {
		hello, err := server.Recv()
		if err != nil {
			return err
		}
		log.Info().Msg(hello.Name)
		fmt.Println(server.Send(hello))
	}
}

func (id *Id) UploadDownload(ctx context.Context, request *gidpb.UploadFileRequest) (*httpbody.HttpBody, error) {
	log.Info().Msg(request.Filename)
	return request.File, nil
}

func (id *Id) Chat(server gidpb.Id_ChatServer) error {
	for {
		hello, err := server.Recv()
		if err != nil {
			return err
		}
		log.Info().Msg(hello.Name)
		server.Send(hello)
	}
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
		time.Sleep(time.Second)
	}
	return nil
}

func (id *Id) Middlewares() []lava.Middleware {
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

func (id *Id) ServiceDesc() *grpc.ServiceDesc {
	return &gidpb.Id_ServiceDesc
}

func New(cron *scheduler.Scheduler, metric metrics.Metric, log log.Logger, service *gid_client.Service) lava.GrpcRouter {
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
	//id.cron.Every("test_gid", time.Second*10, func(ctx context.Context, name string) error {
	//	fmt.Println("test cron every")
	//
	//	rsp, err := id.service.Types(ctx, &gidpb.TypesRequest{})
	//	if err != nil {
	//		return err
	//	}
	//
	//	id.log.Info(ctx).Any("data", rsp.Types).Msg("Types")
	//
	//	defer recovery.Exit()
	//	rsp1 := id.service.Do(ctx, resty.NewRequest(typesReq))
	//	if rsp1.IsErr() {
	//		return rsp1.Err()
	//	}
	//
	//	id.log.Info(ctx).Any("data", string(rsp1.Unwrap().Body())).Msg("Types http")
	//
	//	return nil
	//})
}

func (id *Id) Generate(ctx context.Context, req *gidpb.GenerateRequest) (*gidpb.GenerateResponse, error) {
	//return nil, errors.New("hello error")
	//return nil, errors.NewCodeErr(&errorpb.ErrCode{
	//	StatusCode: errorpb.Code_Internal,
	//	Code:       2222,
	//	Message:    "ddd",
	//	Name:       "dd.sss",
	//}, gidpb.ErrCodeIDGenerateFailed)

	log.Info().Any("type", req.Type.String()).Msg("request")
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
	fmt.Println(req.String() + "\n\n\n\n")
	rsp := new(gidpb.TypesResponse)
	rsp.Types = []string{
		"uuid",
		"shortid",
		"snowflake",
		"bigflake",
	}
	return rsp, nil
}
