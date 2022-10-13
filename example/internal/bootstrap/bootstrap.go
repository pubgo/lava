package bootstrap

import (
	"fmt"

	"github.com/pubgo/dix/di"
	"github.com/pubgo/lava/clients/orm"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/metric/drivers/prometheus"

	"github.com/pubgo/lava/example/internal/handlers/gidrpc"
	"github.com/pubgo/lava/example/internal/handlers/testapi"
	"github.com/pubgo/lava/example/internal/services/casbinservice"
	"github.com/pubgo/lava/example/internal/services/gidsrv"
	"github.com/pubgo/lava/example/internal/services/menuservice"
)

func Init() {
	di.Provide(func(c config.Config) Config {
		var cc = config.Decode[Config](c)
		fmt.Printf("%#v\n", cc)
		return cc
	})

	di.Provide(prometheus.New)
	di.Provide(gidsrv.NewClient)
	di.Provide(gidrpc.New)
	di.Provide(gidsrv.New)
	di.Provide(menuservice.New)
	di.Provide(casbinservice.New)
	di.Provide(orm.New)
	di.Provide(testapi.New)
}
