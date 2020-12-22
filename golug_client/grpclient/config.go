package grpclient

import "time"

var Name = "grpc_client"
var cfg = make(map[string]ClientCfg)

// WithContextDialer
type ClientCfg struct {
	Insecure              bool          `json:"insecure"`
	Block                 bool          `json:"block"`
	IdleNum               uint32        `json:"idle_num"`
	WriteBufferSize       int           `json:"write_buffer_size"`
	ReadBufferSize        int           `json:"read_buffer_size"`
	InitialWindowSize     int32         `json:"initial_window_size"`
	InitialConnWindowSize int32         `json:"initial_conn_window_size"`
	MaxRecvMsgSize        int           `json:"max_recv_msg_size"`
	MaxDelay              time.Duration `json:"max_delay"`
	WithProxy             bool          `json:"with_proxy"`
	KeepaliveParams       struct {
		MaxConnectionAge      string `json:"max_connection_age"`
		MaxConnectionAgeGrace string `json:"max_connection_age_grace"`
		MaxConnectionIdle     string `json:"max_connection_idle"`
		Time                  string `json:"time"`
		Timeout               string `json:"timeout"`
	} `json:"keepalive_params"`
}

func GetCfg() map[string]ClientCfg {
	return cfg
}

func GetDefaultCfg() ClientCfg {
	return ClientCfg{}
}
