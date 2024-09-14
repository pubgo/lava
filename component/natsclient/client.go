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

	logger log.Logger
}

func New(p Param) *Client {
	logger := p.Logger.WithName("nats-client")

	nc := assert.Must1(nats.Connect(p.Cfg.Url, func(o *nats.Options) error {
		o.AllowReconnect = true
		o.Name = fmt.Sprintf("%s/%s/%s", running.Hostname, running.Project, running.InstanceID)
		return nil
	}))

	nc.SetDisconnectErrHandler(func(nc *nats.Conn, err error) {
		logger.Err(err).Msg("nats disconnect")
	})
	nc.SetReconnectHandler(func(nc *nats.Conn) {
		logger.Info().Bool("is_connected", nc.IsConnected()).Msg("nats reconnect")
	})
	nc.SetClosedHandler(func(nc *nats.Conn) {
		logger.Info().Bool("is_closed", nc.IsClosed()).Msg("nats closed")
	})
	nc.SetErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
		logger.Err(err).Msg("nats error")
	})
	nc.SetDiscoveredServersHandler(func(nc *nats.Conn) {
		logger.Info().Bool("is_connected", nc.IsConnected()).Msg("nats discovered")
	})

	log.Info().Bool("is_connected", nc.IsConnected()).Msg("nats connection ...")

	p.Lc.BeforeStop(func() { nc.Close() })

	return &Client{Param: p, logger: logger, Conn: nc}
}
