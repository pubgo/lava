package grpcbuilder

import (
	"time"

	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/lava/pkg/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// default config: google.golang.org/grpc/internal/transport/defaults.go

type KeepaliveParams struct {
	MaxConnectionAge      time.Duration `yaml:"max_connection_age"`
	MaxConnectionAgeGrace time.Duration `yaml:"max_connection_age_grace"`
	MaxConnectionIdle     time.Duration `yaml:"max_connection_idle"`
	Time                  time.Duration `yaml:"time"`
	Timeout               time.Duration `yaml:"timeout"`
}

func (t *KeepaliveParams) ToOpts() grpc.ServerOption {
	return grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionAge:      t.MaxConnectionAge,
		MaxConnectionAgeGrace: t.MaxConnectionAgeGrace,
		MaxConnectionIdle:     t.MaxConnectionIdle,
		Time:                  t.Time,
		Timeout:               t.Timeout,
	})
}

type KeepalivePolicy struct {
	MinTime             time.Duration `yaml:"min_time"`
	PermitWithoutStream bool          `yaml:"permit_without_stream"`
}

func (t *KeepalivePolicy) ToOpts() grpc.ServerOption {
	return grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		MinTime:             t.MinTime,
		PermitWithoutStream: t.PermitWithoutStream,
	})
}

type Config struct {
	Codec                 string           `yaml:"codec"`
	ConnectionTimeout     time.Duration    `yaml:"connection_timeout"`
	Cp                    string           `yaml:"cp"`
	Creds                 string           `yaml:"creds"`
	Dc                    string           `yaml:"dc"`
	HeaderTableSize       int64            `yaml:"header_table_size"`
	InitialConnWindowSize int64            `yaml:"initial_conn_window_size"`
	InitialWindowSize     int64            `yaml:"initial_window_size"`
	KeepaliveParams       *KeepaliveParams `yaml:"keepalive_params"`
	KeepalivePolicy       *KeepalivePolicy `yaml:"keepalive_policy"`
	MaxConcurrentStreams  int64            `yaml:"max_concurrent_streams"`
	MaxHeaderListSize     int64            `yaml:"max_header_list_size"`
	MaxRecvMsgSize        int              `yaml:"max_recv_msg_size"`
	MaxSendMsgSize        int              `yaml:"max_send_msg_size"`
	ReadBufferSize        int64            `yaml:"read_buffer_size"`
	WriteBufferSize       int64            `yaml:"write_buffer_size"`
}

func (t *Config) Build(opts ...grpc.ServerOption) (r result.Result[*grpc.Server]) {
	defer result.RecoveryErr(&r)

	if t.KeepalivePolicy != nil {
		opts = append(opts, t.KeepalivePolicy.ToOpts())
	}

	if t.KeepaliveParams != nil {
		opts = append(opts, t.KeepaliveParams.ToOpts())
	}

	srv := grpc.NewServer(opts...)

	grpcutil.EnableReflection(srv)
	grpcutil.EnableHealth("", srv)
	grpcutil.EnableDebug(srv)
	return r.WithValue(srv)
}

func GetDefaultCfg() *Config {
	return &Config{
		ConnectionTimeout: 120 * time.Second,
	}
}
