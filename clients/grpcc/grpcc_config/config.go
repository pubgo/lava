package grpcc_config

import (
	"github.com/pubgo/lava/core/middleware"
	"time"

	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc/grpcc_resolver"
)

const (
	Name = "grpcc"

	// DefaultTimeout 默认的连接超时时间
	DefaultTimeout     = 2 * time.Second
	DefaultContentType = "application/grpc"
)

var defaultOpts = []grpc.DialOption{grpc.WithDefaultServiceConfig(`{}`)}

// Cfg ...
type Cfg struct {
	Client      *ClientCfg                       `yaml:"client"`
	Addr        string                           `yaml:"addr"`
	Scheme      string                           `yaml:"scheme"`
	Registry    string                           `yaml:"registry"`
	Alias       string                           `yaml:"alias"`
	Middlewares map[string]middleware.Middleware `yaml:"-"`
}

func (t Cfg) Check() error { return nil }

func DefaultCfg() *Cfg {
	var cfg = &Cfg{
		Scheme: grpcc_resolver.DiscovScheme,
		Client: &ClientCfg{
			Insecure:             true,
			Block:                true,
			DefaultServiceConfig: `{"loadBalancingConfig": [{"p2c":{}}]}`,
			DialTimeout:          time.Minute,
			Timeout:              DefaultTimeout,
			MaxHeaderListSize:    1024 * 4,
			MaxRecvMsgSize:       1024 * 1024 * 4,
			ClientParameters: clientParameters{
				PermitWithoutStream: true,             // send pings even without active streams
				Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
				Timeout:             2 * time.Second,  // wait 2 second for ping ack before considering the connection dead
			},
			ConnectParams: connectParams{
				Backoff: backoffConfig{
					Multiplier: 1.6,
					Jitter:     0.2,
					BaseDelay:  1.0 * time.Second,
					MaxDelay:   120 * time.Second,
				},
			},
			Call: callParameters{
				MaxCallRecvMsgSize: 1024 * 1024 * 4,
				// DefaultMaxSendMsgSize maximum message that Srv can send (4 MB).
				MaxCallSendMsgSize: 1024 * 1024 * 4,
			},
		},
	}
	return cfg
}
