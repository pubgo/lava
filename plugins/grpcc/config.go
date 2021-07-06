package grpcc

import (
	"context"
	grpcTracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"time"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/keepalive"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/pkg/typex"
	p2c2 "github.com/pubgo/lug/plugins/grpcc/balancer/p2c"
	"github.com/pubgo/lug/plugins/grpcc/balancer/resolver"
)

const (
	Name = "grpcc"

	// DefaultTimeout 默认的连接超时时间
	DefaultTimeout = 3 * time.Second
)

var configMap typex.SMap

type callParameters struct {
	Header                map[string]string `json:"header"`
	Trailer               map[string]string `json:"trailer"`
	WaitForReady          bool              `json:"wait_for_ready"`
	FailFast              bool              `json:"fail_fast"`
	MaxCallRecvMsgSize    int               `json:"max_call_recv_msg_size"`
	MaxCallSendMsgSize    int               `json:"max_call_send_msg_size"`
	UseCompressor         string            `json:"use_compressor"`
	CallContentSubtype    string            `json:"call_content_subtype"`
	ForceCodec            string            `json:"force_codec"`
	MaxRetryRPCBufferSize int               `json:"max_retry_rpc_buffer_size"`
}

type clientParameters struct {
	PermitWithoutStream bool          `json:"permit_without_stream"`
	Time                time.Duration `json:"time"`
	Timeout             time.Duration `json:"timeout"`
}

func (t clientParameters) toClientParameters() keepalive.ClientParameters {
	return keepalive.ClientParameters{
		PermitWithoutStream: t.PermitWithoutStream,
		Time:                t.Time,
		Timeout:             t.Timeout,
	}
}

// backoffConfig defines the configuration options for backoff.
type backoffConfig struct {
	// BaseDelay is the amount of time to backoff after the first failure.
	BaseDelay time.Duration `json:"base_delay"`
	// Multiplier is the factor with which to multiply backoffs after a
	// failed retry. Should ideally be greater than 1.
	Multiplier float64 `json:"multiplier"`
	// Jitter is the factor with which backoffs are randomized.
	Jitter float64 `json:"jitter"`
	// MaxDelay is the upper bound of backoff delay.
	MaxDelay time.Duration `json:"max_delay"`
}

type connectParams struct {
	// Backoff specifies the configuration options for connection backoff.
	Backoff backoffConfig `json:"backoff"`
	// MinConnectTimeout is the minimum amount of time we are willing to give a
	// connection to complete.
	MinConnectTimeout time.Duration `json:"min_connect_timeout"`
}

func (t connectParams) toConnectParams() grpc.ConnectParams {
	return grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  t.Backoff.BaseDelay,
			Multiplier: t.Backoff.Multiplier,
			Jitter:     t.Backoff.Jitter,
			MaxDelay:   t.Backoff.MaxDelay,
		},
		MinConnectTimeout: t.MinConnectTimeout,
	}
}

// WithContextDialer
type Cfg struct {
	Registry             string           `json:"registry"`
	MaxMsgSize           int              `json:"max_msg_size"`
	Codec                string           `json:"codec"`
	Compressor           string           `json:"compressor"`
	Decompressor         string           `json:"decompressor"`
	Balancer             string           `json:"balancer"`
	BackoffMaxDelay      time.Duration    `json:"backoff_max_delay"`
	Timeout              time.Duration    `json:"timeout"`
	DialTimeout          time.Duration    `json:"dial_timeout"`
	MaxDelay             time.Duration    `json:"max_delay"`
	UserAgent            string           `json:"user_agent"`
	Authority            string           `json:"authority"`
	ChannelzParentID     int64            `json:"channelz_parent_id"`
	DisableServiceConfig bool             `json:"disable_service_config"`
	DefaultServiceConfig string           `json:"default_service_config"`
	DisableRetry         bool             `json:"disable_retry"`
	MaxHeaderListSize    uint32           `json:"max_header_list_size"`
	DisableHealthCheck   bool             `json:"disable_health_check"`
	BalancerName         string           `json:"balancer_name"`
	Insecure             bool             `json:"insecure"`
	Block                bool             `json:"block"`
	IdleNum              uint32           `json:"idle_num"`
	WriteBuffer          int              `json:"write_buffer"`
	ReadBuffer           int              `json:"read_buffer"`
	WindowSize           int32            `json:"window_size"`
	ConnWindowSize       int32            `json:"conn_window_size"`
	MaxRecvMsgSize       int              `json:"max_recv_msg_size"`
	NoProxy              bool             `json:"no_proxy"`
	Proxy                bool             `json:"proxy"`
	ConnectParams        connectParams    `json:"connect_params"`
	ClientParameters     clientParameters `json:"client_parameters"`
	Call                 callParameters   `json:"call"`
}

func (t Cfg) Build(target string, opts ...grpc.DialOption) (_ *grpc.ClientConn, err error) {
	defer xerror.RespErr(&err)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, buildTarget(target), append(t.ToOpts(), opts...)...)
	return conn, xerror.WrapF(err, "DialContext error, target:%s\n", target)
}

func (t Cfg) BuildDirect(target string, opts ...grpc.DialOption) (_ *grpc.ClientConn, err error) {
	defer xerror.RespErr(&err)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	target = resolver.BuildDirectTarget(target)
	opts = append(opts, grpc.WithDefaultServiceConfig(`{}`))
	conn, err := grpc.DialContext(ctx, target, append(t.ToOpts(), opts...)...)
	return conn, xerror.WrapF(err, "DialContext error, target:%s\n", target)
}

func (t Cfg) ToOpts() []grpc.DialOption {
	var opts = defaultDialOpts

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

	// Fail right away
	opts = append(opts, grpc.FailOnNonTempDialError(true))
	opts = append(opts, grpc.WithReadBufferSize(t.ReadBuffer))
	opts = append(opts, grpc.WithWriteBufferSize(t.WriteBuffer))
	opts = append(opts, grpc.WithInitialWindowSize(t.WindowSize))
	opts = append(opts, grpc.WithInitialConnWindowSize(t.ConnWindowSize))
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
	grpc.WithDefaultServiceConfig(`{}`),
	grpc.WithChainUnaryInterceptor(
		grpcTracing.UnaryClientInterceptor(),
	),
	grpc.WithChainStreamInterceptor(
		grpcTracing.StreamClientInterceptor(),
	),
}

func GetCfg(name string) Cfg {
	if configMap.Has(name) {
		return configMap.Get(name).(Cfg)
	} else if configMap.Has(consts.Default) {
		return configMap.Get(consts.Default).(Cfg)
	} else {
		return GetDefaultCfg()
	}
}

func defaultDialOption(_ string) []grpc.DialOption { return GetDefaultCfg().ToOpts() }
func GetDefaultCfg() Cfg {
	return Cfg{
		Insecure:     true,
		Block:        true,
		BalancerName: p2c2.Name,
		DialTimeout:  2 * time.Second,

		// DefaultMaxRecvMsgSize maximum message that client can receive (4 MB).
		MaxRecvMsgSize: 1024 * 1024 * 4,
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
			// DefaultMaxSendMsgSize maximum message that client can send (4 MB).
			MaxCallSendMsgSize: 1024 * 1024 * 4,
		},
	}
}
