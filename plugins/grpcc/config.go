package grpcc

import (
	"context"
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
	"github.com/pubgo/lug/registry"
	"github.com/pubgo/lug/types"
)

const (
	Name = "grpcc"

	// DefaultTimeout 默认的连接超时时间
	DefaultTimeout     = 3 * time.Second
	defaultContentType = "application/grpc"
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

// Cfg ...
type Cfg struct {
	Registry             string        `json:"registry"`
	MaxMsgSize           int           `json:"max_msg_size"`
	Codec                string        `json:"codec"`
	Compressor           string        `json:"compressor"`
	Decompressor         string        `json:"decompressor"`
	Balancer             string        `json:"balancer"`
	BackoffMaxDelay      time.Duration `json:"backoff_max_delay"`
	Timeout              time.Duration `json:"timeout"`
	DialTimeout          time.Duration `json:"dial_timeout"`
	MaxDelay             time.Duration `json:"max_delay"`
	UserAgent            string        `json:"user_agent"`
	Authority            string        `json:"authority"`
	ChannelzParentID     int64         `json:"channelz_parent_id"`
	DisableServiceConfig bool          `json:"disable_service_config"`
	DefaultServiceConfig string        `json:"default_service_config"`
	DisableRetry         bool          `json:"disable_retry"`
	MaxHeaderListSize    uint32        `json:"max_header_list_size"`
	DisableHealthCheck   bool          `json:"disable_health_check"`
	BalancerName         string        `json:"balancer_name"`
	Insecure             bool          `json:"insecure"`
	Block                bool          `json:"block"`
	IdleNum              uint32        `json:"idle_num"`
	WriteBuffer          int           `json:"write_buffer"`
	ReadBuffer           int           `json:"read_buffer"`
	WindowSize           int32         `json:"window_size"`
	ConnWindowSize       int32         `json:"conn_window_size"`

	// MaxRecvMsgSize maximum message that client can receive (4 MB).
	MaxRecvMsgSize     int                            `json:"max_recv_msg_size"`
	NoProxy            bool                           `json:"no_proxy"`
	Proxy              bool                           `json:"proxy"`
	ConnectParams      connectParams                  `json:"connect_params"`
	ClientParameters   clientParameters               `json:"client_parameters"`
	Call               callParameters                 `json:"call"`
	Middlewares        []types.Middleware             `json:"-"`
	DialOptions        []grpc.DialOption              `json:"-"`
	UnaryInterceptors  []grpc.UnaryClientInterceptor  `json:"-"`
	StreamInterceptors []grpc.StreamClientInterceptor `json:"-"`
}

func (t Cfg) Build(target string) (conn *grpc.ClientConn, gErr error) {
	defer xerror.RespErr(&gErr)

	// 服务发现模式, 检查注册中心是否初始化完成
	xerror.Assert(registry.Default() == nil, "please init registry")
	target = resolver.BuildDiscovTarget(target, registry.Default().String())

	ctx, cancel := context.WithTimeout(context.Background(), t.DialTimeout)
	defer cancel()
	//
	conn, gErr = grpc.DialContext(ctx, target, append(t.ToOpts(), t.DialOptions...)...)
	return conn, xerror.WrapF(gErr, "DialContext error, target:%s\n", target)
}

func (t Cfg) BuildDirect(target string) (conn *grpc.ClientConn, gErr error) {
	defer xerror.RespErr(&gErr)

	// 直连模式,target=>localhost:8080,localhost:8081,localhost:8082
	target = resolver.BuildDirectTarget(target)

	ctx, cancel := context.WithTimeout(context.Background(), t.DialTimeout)
	defer cancel()

	conn, gErr = grpc.DialContext(ctx, target, append(t.ToOpts(), t.DialOptions...)...)
	return conn, xerror.WrapF(gErr, "DialContext error, target:%s", target)
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

	if t.ReadBuffer != 0 {
		opts = append(opts, grpc.WithReadBufferSize(t.ReadBuffer))
	}

	if t.WriteBuffer != 0 {
		opts = append(opts, grpc.WithWriteBufferSize(t.WriteBuffer))
	}

	if t.WindowSize != 0 {
		opts = append(opts, grpc.WithInitialWindowSize(t.WindowSize))
	}

	if t.ConnWindowSize != 0 {
		opts = append(opts, grpc.WithInitialConnWindowSize(t.ConnWindowSize))
	}

	if t.UserAgent != "" {
		opts = append(opts, grpc.WithUserAgent(t.UserAgent))
	}

	if t.Authority != "" {
		opts = append(opts, grpc.WithAuthority(t.Authority))
	}

	if t.DefaultServiceConfig != "" {
		opts = append(opts, grpc.WithDefaultServiceConfig(t.DefaultServiceConfig))
	}

	if t.MaxHeaderListSize != 0 {
		opts = append(opts, grpc.WithMaxHeaderListSize(t.MaxHeaderListSize))
	}

	if t.ChannelzParentID != 0 {
		opts = append(opts, grpc.WithChannelzParentID(t.ChannelzParentID))
	}

	var cos []grpc.CallOption
	if t.MaxRecvMsgSize != 0 {
		cos = append(cos, grpc.MaxCallRecvMsgSize(t.MaxRecvMsgSize))
	}

	if t.Codec != "" {
		cos = append(cos, grpc.ForceCodec(encoding.GetCodec(t.Codec)))
	}

	if t.Compressor != "" {
		cos = append(cos, grpc.UseCompressor(t.Compressor))
	}

	opts = append(opts, grpc.WithDefaultCallOptions(cos...))
	opts = append(opts, grpc.FailOnNonTempDialError(true))
	opts = append(opts, grpc.WithKeepaliveParams(t.ClientParameters.toClientParameters()))
	opts = append(opts, grpc.WithConnectParams(t.ConnectParams.toConnectParams()))

	var unaryInterceptors = append([]grpc.UnaryClientInterceptor{unaryInterceptor(t.Middlewares)}, t.UnaryInterceptors...)
	opts = append(opts, grpc.WithChainUnaryInterceptor(unaryInterceptors...))

	var streamInterceptors = append([]grpc.StreamClientInterceptor{streamInterceptor(t.Middlewares)}, t.StreamInterceptors...)
	opts = append(opts, grpc.WithChainStreamInterceptor(streamInterceptors...))
	return opts
}

var defaultDialOpts = []grpc.DialOption{grpc.WithDefaultServiceConfig(`{}`)}

func GetCfg(name string) *Cfg {
	if configMap.Has(name) {
		return configMap.Get(name).(*Cfg)
	}

	if configMap.Has(consts.Default) {
		return configMap.Get(consts.Default).(*Cfg)
	}

	return GetDefaultCfg()
}

func GetDefaultCfg(opts ...func(cfg *Cfg)) *Cfg {
	var cfg = Cfg{
		Insecure:          true,
		Block:             true,
		BalancerName:      p2c2.Name,
		DialTimeout:       time.Minute,
		Timeout:           DefaultTimeout,
		MaxHeaderListSize: 1024 * 4,
		MaxRecvMsgSize:    1024 * 1024 * 4,
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

	for i := range opts {
		opts[i](&cfg)
	}
	return &cfg
}
