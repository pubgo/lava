package bootstrap

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/cmds/builder"

	"github.com/pubgo/lava/internal/example/grpc/handlers/gid_handler"
	"github.com/pubgo/lava/internal/example/grpc/services/gid_client"
)

func Main() {
	defer recovery.Exit()

	var di = builder.NewDix(dix.WithValuesNull())
	di.Provide(config.Load[Config])

	di.Provide(gid_handler.New)
	di.Provide(gid_client.New)

	builder.Run(di)
}
