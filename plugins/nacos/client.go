package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/xerror"
)

func Get(names ...string) *Client {
	var name = lavax.GetDefault(names...)
	var cfg, ok = cfgMap[name]
	if ok {
		return cfg.c
	}
	return nil
}

type Client struct {
	srv naming_client.INamingClient
	cfg config_client.IConfigClient
}

func (c Client) GetCfg() config_client.IConfigClient {
	xerror.Assert(c.cfg == nil, "please init config client")
	return c.cfg
}

func (c Client) GetRegistry() naming_client.INamingClient {
	xerror.Assert(c.srv == nil, "please init naming client")
	return c.srv
}
