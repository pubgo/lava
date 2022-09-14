package bootstrap

import (
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/example/internal/services/casbinservice"
	"github.com/pubgo/lava/example/internal/services/menuservice"
)

type Config struct {
	Casbin *casbinservice.Config `yaml:"casbin"`
	Menu   *menuservice.Config   `yaml:"menu"`
	Db     *orm.Cfg              `yaml:"orm"`
}
