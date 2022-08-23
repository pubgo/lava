package bootstrap

import (
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/example/internal/services/casbinservice"
	"github.com/pubgo/lava/example/internal/services/menuservice"
)

type Config struct {
	Casbin *casbinservice.Config `json:"casbin"`
	Menu   *menuservice.Config   `json:"menu"`
	Db     *orm.Cfg              `json:"orm"`
}
