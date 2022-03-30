package rs_manager

import (
	"github.com/pubgo/lava/encoding"
	"sync"

	"github.com/rsocket/rsocket-go"
)

var services sync.Map

var _ rsocket.RSocket = (*Service)(nil)

type Service struct {
	rsocket.RSocket
	Kind        string
	Name        string
	Id          string
	Version     string
	Environment string
	HostName    string            `json:"hostName"`
	Status      string            `json:"status"`
	Metadata    map[string]string `json:"metadata"`
	MetaCodec   encoding.Codec
	DataCodec   encoding.Codec
}

func AddService(srv *Service) {

}

func GetService(name string) *Service {

}

func DelService(srv *Service) *Service {

}
