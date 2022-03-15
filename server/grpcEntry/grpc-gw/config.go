package grpc_gw

import (
	"time"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

const DefaultTimeout = time.Second * 2

func init() {
	gw.DefaultContextTimeout = DefaultTimeout
}

type ServeMux = gw.ServeMux

type Cfg struct {
	Timeout time.Duration `json:"timeout"`
}

func DefaultCfg() *Cfg {
	return &Cfg{
		Timeout: time.Second * 2,
	}
}
