package gidrpc

import (
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/example/gen/proto/gidpb"
	"math/rand"
)

func New(cron *scheduler.Scheduler, metric metric.Metric) gidpb.IdServer {
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
		m:         metric,
		snowflake: sf,
		bigflake:  bg,
	}
}
