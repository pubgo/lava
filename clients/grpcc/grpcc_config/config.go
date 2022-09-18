package grpcc_config

import (
	"time"

	"github.com/pubgo/lava/clients/grpcc/grpcc_resolver"
	"github.com/pubgo/lava/service"
	"google.golang.org/grpc"
)

const (
	Name = "grpcc"

	// DefaultTimeout 默认的连接超时时间
	DefaultTimeout     = 2 * time.Second
	DefaultContentType = "application/grpc"
)

var defaultOpts = []grpc.DialOption{grpc.WithDefaultServiceConfig(`{
	"loadBalancingConfig": [{"round_robin": {}}],
	"methodConfig": [{
		"name": [{"service": ""}],
		"waitForReady": true,
		"retryPolicy": {
			"MaxAttempts": 5,
			"InitialBackoff": "0.1s",
			"MaxBackoff": "5s",
			"BackoffMultiplier": 2,
			"RetryableStatusCodes": ["UNAVAILABLE"]
		}
	}]
}`)}

// Cfg ...
type Cfg struct {
	Client      *ClientCfg           `yaml:"client"`
	Srv         string               `yaml:"srv"`
	Addr        string               `yaml:"addr"`
	Scheme      string               `yaml:"scheme"`
	Alias       string               `yaml:"alias"`
	Middlewares []service.Middleware `yaml:"-"`
}

func (t Cfg) Check() error { return nil }

func DefaultCfg() *Cfg {
	var cfg = &Cfg{
		Scheme: grpcc_resolver.DiscovScheme,
		Client: &ClientCfg{
			Insecure: true,
			Block:    true,
			// refer: https://github.com/grpc/grpc/blob/master/doc/service_config.md
			// refer: https://github.com/grpc/grpc-proto/blob/d653c6d98105b2af937511aa6e46610c7e677e6e/grpc/service_config/service_config.proto#L632
			DialTimeout:       time.Minute,
			Timeout:           DefaultTimeout,
			MaxHeaderListSize: 1024 * 4,
			MaxRecvMsgSize:    1024 * 1024 * 4,
			// refer: https://github.com/grpc/grpc-go/blob/master/examples/features/keepalive/client/main.go
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
