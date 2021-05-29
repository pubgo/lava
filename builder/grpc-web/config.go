package grpcWeb

import (
	fb "github.com/pubgo/lug/builder/fiber"

	"net/http"
)

const Name = "grpc-web"

type Middleware func(w http.ResponseWriter, r *http.Request)

type Cfg struct {
	fb.Cfg
	Prefix string
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Prefix: "",
	}
}
