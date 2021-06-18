package grpcWeb

import (
	"net/http"
)

const Name = "grpc-web"

type Middleware func(w http.ResponseWriter, r *http.Request)

type Cfg struct {
	Prefix string
}

func GetDefaultCfg() *Cfg {
	return &Cfg{
		Prefix: "",
	}
}
