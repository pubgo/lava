package gidsrv

import (
	"context"
	"math/rand"

	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/errors"
	"github.com/pubgo/lava/example/gen/proto/gidpb"
)

func New(cc *Client) Service {
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
		cc:        cc,
		snowflake: sf,
		bigflake:  bg,
	}
}

var err1 = errors.New("id.generate")

var (
	_ gidpb.IdServer = (*Id)(nil)
)

type Id struct {
	cc        *Client
	cron      *scheduler.Scheduler
	snowflake *snowflake.Snowflake
	bigflake  *bigflake.Bigflake
}

func (id *Id) GetTypes() []string {
	rsp, _ := id.cc.Types(context.Background(), new(gidpb.TypesRequest))
	return rsp.Types
}
