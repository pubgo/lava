package grpcs

import (
	"time"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc/keepalive"
)

const (
	DefaultContentType = "application/grpc"
	DefaultMaxMsgSize  = 1024 * 1024 * 4
	DefaultTimeout     = time.Second * 2
)

type KeepaliveParams struct {
	MaxConnectionAge      time.Duration `json:"max_connection_age"`
	MaxConnectionAgeGrace time.Duration `json:"max_connection_age_grace"`
	MaxConnectionIdle     time.Duration `json:"max_connection_idle"`
	Time                  time.Duration `json:"time"`
	Timeout               time.Duration `json:"timeout"`
}

func (t KeepaliveParams) ToCfg() (sp keepalive.ServerParameters) {
	xerror.Panic(merge.Copy(&sp, &t))
	return
}

type KeepalivePolicy struct {
	MinTime             time.Duration `json:"min_time"`
	PermitWithoutStream bool          `json:"permit_without_stream"`
}

type Cfg struct {
	Codec                 string          `json:"codec"`
	ConnectionTimeout     time.Duration   `json:"connection_timeout"`
	Cp                    string          `json:"cp"`
	Creds                 string          `json:"creds"`
	Dc                    string          `json:"dc"`
	HeaderTableSize       int64           `json:"header_table_size"`
	InitialConnWindowSize int64           `json:"initial_conn_window_size"`
	InitialWindowSize    int64           `json:"initial_window_size"`
	KeepaliveParams      KeepaliveParams `json:"keepalive_params"`
	KeepalivePolicy      KeepalivePolicy `json:"keepalive_policy"`
	MaxConcurrentStreams int64           `json:"max_concurrent_streams"`
	MaxHeaderListSize     int64           `json:"max_header_list_size"`
	MaxRecvMsgSize        int             `json:"max_recv_msg_size"`
	MaxSendMsgSize        int             `json:"max_send_msg_size"`
	ReadBufferSize        int64           `json:"read_buffer_size"`
	WriteBufferSize       int64           `json:"write_buffer_size"`
}

func GetDefaultCfg() *Cfg {
	return &Cfg{
		MaxRecvMsgSize:    DefaultMaxMsgSize,
		MaxSendMsgSize:    DefaultMaxMsgSize,
		WriteBufferSize:   32 * 1024,
		ReadBufferSize:    32 * 1024,
		ConnectionTimeout: 120 * time.Second,
		KeepaliveParams: KeepaliveParams{
			MaxConnectionIdle:     30 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
			MaxConnectionAge:      55 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
			MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
			Time:                  10 * time.Second, // Ping the client if it is idle for 5 seconds to ensure the connection is still active
			Timeout:               2 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
		},
		KeepalivePolicy: KeepalivePolicy{
			MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
			PermitWithoutStream: true,            // Allow pings even when there are no active streams
		},
	}
}
