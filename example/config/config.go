package config

import (
	"github.com/pubgo/lava/example/internal/casbin"
	"github.com/pubgo/lava/example/internal/menuservice"
)

type Config struct {
	Casbin *casbin.Config      `json:"casbin"`
	Menu   *menuservice.Config `json:"menu"`
}
