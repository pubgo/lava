package grpclient

import (
	"sync"
	"time"
)

var Name = "grpc_client"
var cfg = make(map[string]ClientCfg)
var connPool sync.Map
var maxConnRef = uint32(50)

const (
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
	PermitWithoutStream bool          `json:"permit_without_stream"`
	Time                time.Duration `json:"time"`
	Timeout             time.Duration `json:"timeout"`
}

// BackoffConfig defines the configuration options for backoff.
type BackoffConfig struct {
	// BaseDelay is the amount of time to backoff after the first failure.
	BaseDelay time.Duration
	// Multiplier is the factor with which to multiply backoffs after a
	// failed retry. Should ideally be greater than 1.
	Multiplier float64
	// Jitter is the factor with which backoffs are randomized.
	Jitter float64
	// MaxDelay is the upper bound of backoff delay.
	MaxDelay time.Duration
}

type ConnectParams struct {
	// Backoff specifies the configuration options for connection backoff.
	Backoff BackoffConfig
	// MinConnectTimeout is the minimum amount of time we are willing to give a
	// connection to complete.
	MinConnectTimeout time.Duration
}

// WithContextDialer
type ClientCfg struct {
	MaxMsgSize int
	// grpc.encoding
	Codec                string
	Compressor           string
	Decompressor         string
	Balancer             string
	BackoffMaxDelay      time.Duration
	Timeout              time.Duration
	DialTimeout          time.Duration
	UserAgent            string
	ConnectParams        ConnectParams
	Authority            string
	ChannelzParentID     int64
	DisableServiceConfig bool
	DefaultServiceConfig string
	DisableRetry         bool
	MaxHeaderListSize    uint32
	DisableHealthCheck   bool
	Insecure             bool          `json:"insecure"`
	Block                bool          `json:"block"`
	IdleNum              uint32        `json:"idle_num"`
	WriteBuffer          int           `json:"write_buffer"`
	ReadBuffer           int           `json:"read_buffer"`
	WindowSize           int32         `json:"window_size"`
	ConnWindowSize       int32         `json:"conn_window_size"`
	MaxRecvMsgSize       int           `json:"max_recv_msg_size"`
	MaxDelay             time.Duration `json:"max_delay"`
	NoProxy              bool
	Proxy                bool             `json:"proxy"`
	ClientParameters     ClientParameters `json:"params"`
	Call                 Call             `json:"call"`
}

func GetCfg() map[string]ClientCfg {
	return cfg
}

func GetDefaultCfg() ClientCfg {
	return ClientCfg{
		DialTimeout: 2 * time.Second,
		// DefaultMaxRecvMsgSize maximum message that client can receive (4 MB).
		MaxRecvMsgSize: 1024 * 1024 * 4,
		ClientParameters: ClientParameters{
			Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			Timeout:             2 * time.Second,  // wait 2 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // send pings even without active streams
		},
		ConnectParams: ConnectParams{
			Backoff: BackoffConfig{
				BaseDelay:  1.0 * time.Second,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   120 * time.Second,
			},
		},
		Call: Call{
			// DefaultMaxSendMsgSize maximum message that client can send (4 MB).
			MaxCallSendMsgSize: 1024 * 1024 * 4,
		},
	}
}
