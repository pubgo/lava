package config

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
)

func init() {
	dix.Provide(func() Config { return newCfg() })
}

func Decode[Cfg any](c Config, name string) map[string]Cfg {
	var cfgMap = make(map[string]Cfg)
	assert.MustF(c.Decode(name, &cfgMap), "config decode failed, name=%s", name)
	return cfgMap
}

func MakeClient[Cfg any, Client any](c Config, name string, callback func(key string, cfg Cfg) Client) map[string]Client {
	var cfgMap = Decode[Cfg](c, name)
	var clientMap = make(map[string]Client)
	for key := range cfgMap {
		clientMap[key] = callback(key, cfgMap[key])
	}
	return clientMap
}
