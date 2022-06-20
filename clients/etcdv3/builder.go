package etcdv3

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lava/config"
)

const Name = "etcdv3"

func init() {
	dix.Provider(func(c config.Config) map[string]*Client {
		return config.MakeClient(c, Name, func(key string, cfg *Cfg) *Client {
			return &Client{Client: cfg.Build()}
		})
	})
}
