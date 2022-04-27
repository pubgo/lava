package running

import (
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/middleware"
)

type Running interface {
	AfterStops(...func())
	BeforeStops(...func())
	AfterStarts(...func())
	BeforeStarts(...func())
	Flags(flags ...cli.Flag)
	Middlewares(middleware.Middleware)
}
