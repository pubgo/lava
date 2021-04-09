package grpcEntry

import (
	"math"
	"time"

	"github.com/google/uuid"
	opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/pubgo/lug/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultClientMaxReceiveMessageSize = 1024 * 1024 * 4
	defaultClientMaxSendMessageSize    = math.MaxInt32
	// http2IOBufSize specifies the buffer size for sending frames.
	defaultWriteBufSize                = 32 * 1024
	defaultReadBufSize                 = 32 * 1024
	defaultServerMaxReceiveMessageSize = 1024 * 1024 * 4
	defaultServerMaxSendMessageSize    = math.MaxInt32
	connectionTimeout                  = 120 * time.Second

	Name = "grpc_entry"

	// The register expiry time
	RegisterTTL = time.Minute
	// The interval on which to register
	RegisterInterval = time.Second * 30

	defaultContentType = "application/grpc"

	defaultUnaryTimeout  = 10 * time.Second
	defaultStreamTimeout = 10 * time.Second
	registryName         = ""

	// DefaultMaxMsgSize define maximum message size that server can send
	// or receive.  Default value is 4MB.
	DefaultMaxMsgSize           = 1024 * 1024 * 4
	DefaultSleepAfterDeregister = time.Second * 2
	// The register expiry time
	DefaultRegisterTTL = time.Minute
	// The interval on which to register
	DefaultRegisterInterval = time.Second * 30
)

var (
	DefaultId = uuid.New().String()

	streamInterceptors = []grpc.StreamServerInterceptor{
		opentracing.StreamServerInterceptor(),
	}

	unaryInterceptors = []grpc.UnaryServerInterceptor{
		opentracing.UnaryServerInterceptor(),
	}
)

func GetDefaultServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(DefaultMaxMsgSize),
		grpc.MaxSendMsgSize(DefaultMaxMsgSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
			PermitWithoutStream: true,            // Allow pings even when there are no active streams
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
			PermitWithoutStream: true,            // Allow pings even when there are no active streams
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     30 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
			MaxConnectionAge:      55 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
			MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
			Time:                  10 * time.Second, // Ping the client if it is idle for 5 seconds to ensure the connection is still active
			Timeout:               2 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
		}),
	}
}

type Cfg struct {
	Gw struct {
		Addr string `json:"addr"`
	} `json:"gw"`
	SleepAfterDeregister  time.Duration `json:"SleepAfterDeregister"`
	RegisterInterval      time.Duration `json:"RegisterInterval"`
	RegisterTTL           string        `json:"register_ttl"`
	Address               string        `json:"address"`
	Advertise             string        `json:"advertise"`
	Codec                 string        `json:"codec"`
	ConnectionTimeout     string        `json:"connection_timeout"`
	Cp                    string        `json:"cp"`
	Creds                 string        `json:"creds"`
	Dc                    string        `json:"dc"`
	HeaderTableSize       int64         `json:"header_table_size"`
	InitialConnWindowSize int64         `json:"initial_conn_window_size"`
	InitialWindowSize     int64         `json:"initial_window_size"`
	KeepaliveParams       struct {
		MaxConnectionAge      string `json:"max_connection_age"`
		MaxConnectionAgeGrace string `json:"max_connection_age_grace"`
		MaxConnectionIdle     string `json:"max_connection_idle"`
		Time                  string `json:"time"`
		Timeout               string `json:"timeout"`
	} `json:"keepalive_params"`
	KeepalivePolicy struct {
		MinTime             string `json:"min_time"`
		PermitWithoutStream bool   `json:"permit_without_stream"`
	} `json:"keepalive_policy"`
	MaxConcurrentStreams  int64 `json:"max_concurrent_streams"`
	MaxHeaderListSize     int64 `json:"max_header_list_size"`
	MaxReceiveMessageSize int   `json:"max_receive_message_size"`
	MaxSendMessageSize    int   `json:"max_send_message_size"`
	ReadBufferSize        int64 `json:"read_buffer_size"`
	WriteBufferSize       int64 `json:"write_buffer_size"`
	registry              registry.Registry
}

var DefaultCfg = Cfg{
	MaxReceiveMessageSize: 1,
	MaxSendMessageSize:    1,
}

const name = `
{
  "write_buffer_size": 1,
  "read_buffer_size": 1,
  "initial_window_size": 1,
  "initial_conn_window_size": 1,
  "keepalive_params": {
    "max_connection_idle": "1s",
    "max_connection_age": "2s",
    "max_connection_age_grace": "2s",
    "time": "1s",
    "timeout": "1s"
  },
  "keepalive_policy": {
    "permit_without_stream": true,
    "min_time": "1s"
  },
  "codec": "json",
  "cp": "gzip",
  "dc": "gzip",
  "max_receive_message_size": 1,
  "max_send_message_size": 1,
  "max_concurrent_streams": 1,
  "creds": "tls",
  "connection_timeout": "2s",
  "max_header_list_size": 2,
  "header_table_size": 1
}
`
