package grpclient

import "time"

var Name = "grpc_client"
var cfg = make(map[string]ClientCfg)

// WithContextDialer
type ClientCfg struct {
	Insecure              bool
	Block                 bool
	IdleNum               uint32
	WriteBufferSize       int
	ReadBufferSize        int
	InitialWindowSize     int32
	InitialConnWindowSize int32
	MaxRecvMsgSize        int
	MaxDelay              time.Duration
	withProxy             bool
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
