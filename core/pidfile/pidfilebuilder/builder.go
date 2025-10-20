package pidfilebuilder

import (
	"context"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/pidfile"
)

func New() lifecycle.Handler {
	return func(lc lifecycle.Lifecycle) {
		lc.AfterStart(func(ctx context.Context) error { return pidfile.Save().GetErr() })
	}
}
