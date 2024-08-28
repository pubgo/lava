package natsclient

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/lava/core/lifecycle"
)

type Param struct {
	Cfg    *Config
	Logger log.Logger
	Lc     lifecycle.Lifecycle
}

type Client struct {
	Param

	*nats.Conn
}

func New(p Param) *Client {
	c := &Client{Param: p}
	c.Logger = c.Logger.WithName("nats")

	nc := assert.Must1(nats.Connect(c.Cfg.Url, func(o *nats.Options) error {
		o.AllowReconnect = true
		o.Name = fmt.Sprintf("%s/%s/%s", running.Hostname, running.Project, running.InstanceID)
		return nil
	}))
	log.Info().Bool("status", nc.IsConnected()).Msg("nats connection ...")

	c.Lc.BeforeStop(func() {
		nc.Close()
	})

	c.Conn = nc

	return c
}
