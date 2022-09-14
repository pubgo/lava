package gidsrv

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/errors"
	"github.com/pubgo/lava/example/gen/proto/gidpb"
	"github.com/pubgo/lava/example/gen/proto/hellopb"
)

func New(cron *scheduler.Scheduler, metric metric.Metric) Service {
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
		//testApiSrv: hellopb.NewTestApiClient(conns["test-grpc"]),
		cron:      cron,
		m:         metric,
		snowflake: sf,
		bigflake:  bg,
	}
}

var err1 = errors.New("id.generate")

var (
	_ gidpb.IdServer = (*Id)(nil)
)

type Id struct {
	testApiSrv hellopb.TestApiClient
	cron       *scheduler.Scheduler
	m          metric.Metric
	snowflake  *snowflake.Snowflake
	bigflake   *bigflake.Bigflake
}

func (id *Id) Init() {
	id.cron.Every("test gid", time.Second*2, func(name string) {
		//id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Counter(name).Inc(1)
		//id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Gauge(name).Update(1)
		id.m.Tagged(metric.Tags{"module": "scheduler"}).Gauge(name).Update(1)
		fmt.Println("test cron every")
	})
}