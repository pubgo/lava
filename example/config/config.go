package config

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/config"

	"github.com/pubgo/lava/example/internal/casbin"
	"github.com/pubgo/lava/example/internal/menuservice"
)

type Config struct {
	Casbin *casbin.Config      `json:"casbin"`
	Menu   *menuservice.Config `json:"menu"`
}

func init() {
	dix.Provider(func(c config.Config) *Config {
		var cfg = new(Config)
		assert.Must(c.Unmarshal(cfg))
		return cfg
	})
}
