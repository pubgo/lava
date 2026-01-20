package main

import (
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/env"
	"github.com/pubgo/funk/v2/recovery"

	"github.com/pubgo/lava/v2/core/lavabuilder"
)

func main() {
	defer recovery.Exit()

	version.SetVersion("v1.0.0")
	version.SetProject("tunnel-gateway")
	config.SetConfigPath("internal/configs/tunnel.yaml")
	env.LoadFiles(".env").Must()

	builder := lavabuilder.New()
	lavabuilder.Run(builder)
}
