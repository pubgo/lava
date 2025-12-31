package main

import (
	"context"

	"github.com/pubgo/funk/v2/env"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/running"

	"github.com/pubgo/lava/v2/cmds/fileservercmd"
	"github.com/pubgo/lava/v2/core/lavabuilder"
	"github.com/pubgo/lava/v2/core/logging/logext/slog"
)

func main() {
	defer recovery.Exit()

	slog.SetLogger(log.GetLogger())
	log.SetEnableChecker(func(ctx context.Context, lvl log.Level, name, message string, fields log.Fields) bool {
		if running.Debug.Value() {
			return true
		}

		if name == "dix" || name == "env" {
			return false
		}
		return true
	})

	env.Reload()

	builder := lavabuilder.New()
	builder.Provide(fileservercmd.New)

	lavabuilder.Run(builder)
}
