package gidrpc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/service"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/example/gen/proto/gidpb"
	"github.com/pubgo/lava/example/internal/services/gidsrv"
)

func New(srv gidsrv.Service, cron *scheduler.Scheduler, m metric.Metric) service.GrpcHandler {
	return &Id{
		m:    m,
		srv:  srv,
		cron: cron,
	}
}

type Id struct {
	cron *scheduler.Scheduler
	srv  gidsrv.Service
	m    metric.Metric
}

func (id *Id) ServiceDesc() *grpc.ServiceDesc {
	return &gidpb.Id_ServiceDesc
}

func (id *Id) TwirpHandler(opts ...interface{}) http.Handler {
	return gidpb.NewIdServer(id, opts...)
}

func (id *Id) Middlewares() []service.Middleware {
	return []service.Middleware{func(next service.HandlerFunc) service.HandlerFunc {
		return func(ctx context.Context, req service.Request, rsp service.Response) error {
			fmt.Println(req.Service(), gidpb.Id_ServiceDesc.ServiceName)

			if req.Service() != gidpb.Id_ServiceDesc.ServiceName {
				return next(ctx, req, rsp)
			}

			return next(ctx, req, rsp)
		}
	}}
}

func (id *Id) Init() {
	id.cron.Every("test gid", time.Second*2, func(name string) {
		//id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Counter(name).Inc(1)
		//id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Gauge(name).Update(1)
		id.m.Tagged(metric.Tags{"module": "scheduler"}).Gauge(name).Update(1)
		fmt.Println("test cron every")
		fmt.Printf("types: %v", id.srv.GetTypes())
	})
}
