package bootstrap

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/cmds/running"
	"github.com/pubgo/lava/core/config"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/core/scheduler"

	"github.com/pubgo/lava/internal/example/grpc/handlers/gidhandler"
)

func Main() {
	defer recovery.Exit()

	di.Provide(config.Load[Config])
	di.Provide(logging.New)
	di.Provide(metric.New)
	di.Provide(scheduler.New)
	di.Provide(gidhandler.New)

	running.Main()
}
