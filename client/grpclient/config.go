package grpclient

import (
	"time"

	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/pubgo/golug/client/grpclient/balancer/p2c"
	"github.com/pubgo/golug/golug_types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/keepalive"
)

const (
	Name = "grpc_client"

	// DefaultTimeout 默认的连接超时时间
	DefaultTimeout = 3 * time.Second

	// DefaultMaxRecvMsgSize maximum message that client can receive
	// (4 MB).
	DefaultMaxRecvMsgSize = 1024 * 1024 * 4

	// DefaultMaxSendMsgSize maximum message that client can send
	// (4 MB).
	DefaultMaxSendMsgSize = 1024 * 1024 * 4

	Timeout = 3 * time.Second

	// 闲时每个连接处理的请求
	requestPerConn = 8
)

type Call struct {
	Header                map[string]string
	Trailer               map[string]string
	WaitForReady          bool
	FailFast              bool
	MaxCallRecvMsgSize    int
	MaxCallSendMsgSize    int
	UseCompressor         string
	CallContentSubtype    string
	ForceCodec            string
	MaxRetryRPCBufferSize int
}

type ClientParameters struct {
	PermitWithoutStream bool                 `json:"permit_without_stream"`
	Time                golug_types.Duration `json:"time"`
	Timeout             golug_types.Duration `json:"timeout"`
}

func (t ClientParameters) toClientParameters() keepalive.ClientParameters {
	return keepalive.ClientParameters{
		PermitWithoutStream: t.PermitWithoutStream,
		Time:                t.Time.Duration,
		Timeout:             t.Timeout.Duration,
	}
}

// BackoffConfig defines the configuration options for backoff.
type BackoffConfig struct {
	// BaseDelay is the amount of time to backoff after the first failure.
	BaseDelay golug_types.Duration
	// Multiplier is the factor with which to multiply backoffs after a
	// failed retry. Should ideally be greater than 1.
	Multiplier float64
	// Jitter is the factor with which backoffs are randomized.
	Jitter float64
	// MaxDelay is the upper bound of backoff delay.
	MaxDelay golug_types.Duration
}

type ConnectParams struct {
	// Backoff specifies the configuration options for connection backoff.
	Backoff BackoffConfig
	// MinConnectTimeout is the minimum amount of time we are willing to give a
	// connection to complete.
	MinConnectTimeout golug_types.Duration
}

func (t ConnectParams) toConnectParams() grpc.ConnectParams {
	return grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  t.Backoff.BaseDelay.Duration,
			Multiplier: t.Backoff.Multiplier,
			Jitter:     t.Backoff.Jitter,
			MaxDelay:   t.Backoff.MaxDelay.Duration,
		},
		MinConnectTimeout: t.MinConnectTimeout.Duration,
	}
}

// WithContextDialer
type Cfg struct {
	Registry             string `json:"registry"`
	MaxMsgSize           int
	Codec                string
	Compressor           string
	Decompressor         string
	Balancer             string
	BackoffMaxDelay      golug_types.Duration
	Timeout              golug_types.Duration
	DialTimeout          golug_types.Duration
	MaxDelay             golug_types.Duration `json:"max_delay"`
	UserAgent            string
	ConnectParams        ConnectParams
	Authority            string
	ChannelzParentID     int64
	DisableServiceConfig bool
	DefaultServiceConfig string
	DisableRetry         bool
	MaxHeaderListSize    uint32
	DisableHealthCheck   bool
	BalancerName         string `json:"balancer_name"`
	Insecure             bool   `json:"insecure"`
	Block                bool   `json:"block"`
	IdleNum              uint32 `json:"idle_num"`
	WriteBuffer          int    `json:"write_buffer"`
	ReadBuffer           int    `json:"read_buffer"`
	WindowSize           int32  `json:"window_size"`
	ConnWindowSize       int32  `json:"conn_window_size"`
	MaxRecvMsgSize       int    `json:"max_recv_msg_size"`
	NoProxy              bool
	Proxy                bool             `json:"proxy"`
	ClientParameters     ClientParameters `json:"params"`
	Call                 Call             `json:"call"`
}

func (t Cfg) toOptions() []grpc.DialOption {
	var opts []grpc.DialOption

	if t.Insecure {
		opts = append(opts, grpc.WithInsecure())
	}

	if t.Block {
		opts = append(opts, grpc.WithBlock())
	}

	if t.BalancerName != "" {
		opts = append(opts, grpc.WithBalancerName(t.BalancerName))
	}

	if t.Proxy {
		opts = append(opts, grpc.WithNoProxy())
	}

	if t.DisableServiceConfig {
		opts = append(opts, grpc.WithDisableServiceConfig())
	}

	if t.DisableRetry {
		opts = append(opts, grpc.WithDisableRetry())
	}

	if t.DisableHealthCheck {
		opts = append(opts, grpc.WithDisableHealthCheck())
	}

	opts = append(opts, grpc.WithReadBufferSize(t.ReadBuffer))
	opts = append(opts, grpc.WithWriteBufferSize(t.WriteBuffer))
	opts = append(opts, grpc.WithInitialWindowSize(t.WindowSize))
	opts = append(opts, grpc.WithInitialConnWindowSize(t.ConnWindowSize))
	opts = append(opts, grpc.WithBalancerName(t.BalancerName))
	opts = append(opts, grpc.WithUserAgent(t.UserAgent))
	opts = append(opts, grpc.WithAuthority(t.Authority))
	opts = append(opts, grpc.WithDefaultServiceConfig(t.DefaultServiceConfig))
	opts = append(opts, grpc.WithMaxHeaderListSize(t.MaxHeaderListSize))
	opts = append(opts, grpc.WithChannelzParentID(t.ChannelzParentID))
	opts = append(opts, grpc.WithKeepaliveParams(t.ClientParameters.toClientParameters()))
	opts = append(opts, grpc.WithConnectParams(t.ConnectParams.toConnectParams()))
	opts = append(opts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(t.MaxMsgSize),
		grpc.ForceCodec(encoding.GetCodec(t.Codec)),
		grpc.UseCompressor(t.Compressor),
	))

	return opts
}

var defaultDialOpts = []grpc.DialOption{
	grpc.WithInsecure(),
	grpc.WithBlock(),
	grpc.WithBalancerName(p2c.Name), //nolint:staticcheck
	grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             2 * time.Second,  // wait 2 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}),
	grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(DefaultMaxRecvMsgSize),
		grpc.MaxCallSendMsgSize(DefaultMaxSendMsgSize)),
	grpc.WithChainUnaryInterceptor(
		grpc_opentracing.UnaryClientInterceptor(),
	),
	grpc.WithChainStreamInterceptor(
		grpc_opentracing.StreamClientInterceptor(),
	),
}

func GetDefaultCfg() Cfg {
	return Cfg{
		DialTimeout: golug_types.NewDuration(2 * time.Second),
		// DefaultMaxRecvMsgSize maximum message that client can receive (4 MB).
		MaxRecvMsgSize: 1024 * 1024 * 4,
		ClientParameters: ClientParameters{
			Time:                golug_types.NewDuration(10 * time.Second), // send pings every 10 seconds if there is no activity
			Timeout:             golug_types.NewDuration(2 * time.Second),  // wait 2 second for ping ack before considering the connection dead
			PermitWithoutStream: true,                                      // send pings even without active streams
		},
		ConnectParams: ConnectParams{
			Backoff: BackoffConfig{
				BaseDelay:  golug_types.NewDuration(1.0 * time.Second),
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   golug_types.NewDuration(120 * time.Second),
			},
		},
		Call: Call{
			// DefaultMaxSendMsgSize maximum message that client can send (4 MB).
			MaxCallSendMsgSize: 1024 * 1024 * 4,
		},
	}
}
