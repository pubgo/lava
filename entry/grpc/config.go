package grpc

import (
	grpcGw "github.com/pubgo/lug/builder/grpc-gw"
	grpcWeb "github.com/pubgo/lug/builder/grpc-web"
	"github.com/pubgo/lug/builder/grpcs"

	"time"
)

type Cfg struct {
	Srv                  grpcs.Cfg     `json:"grpc"`
	Gw                   grpcGw.Cfg    `json:"gw"`
	Web                  grpcWeb.Cfg   `json:"web"`
	SleepAfterDeregister time.Duration `json:"sleepAfterDeregister"`
	RegisterInterval     time.Duration `json:"registerInterval"`
	RegisterTTL          time.Duration `json:"register_ttl"`
	Address              string        `json:"address"`
	Advertise            string        `json:"advertise"`
	hostname             string
	id                   string
	name                 string
}
