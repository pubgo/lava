package main

import (
	"github.com/pubgo/dix/dixinternal"
	"github.com/pubgo/funk/cmds/configcmd"
	"github.com/pubgo/funk/cmds/envcmd"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/core/lavabuilder"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/metrics"
	"github.com/rs/zerolog"
)

type Config struct {
	metrics.MetricConfigLoader `yaml:",inline"`
	logging.LogConfigLoader    `yaml:",inline"`
}

func main() {
	defer recovery.Exit()

	dixinternal.SetLog(func(logger log.Logger) log.Logger {
		return logger.WithLevel(zerolog.InfoLevel)
	})

	builder := lavabuilder.New()
	builder.Provide(config.Load[Config])
	builder.Provide(envcmd.New)
	builder.Provide(configcmd.New[Config])

	lavabuilder.Run(builder)
}
