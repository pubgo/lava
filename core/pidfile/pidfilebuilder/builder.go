package pidfilebuilder

import (
	"context"
	"path/filepath"

	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/pathutil"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/pidfile"
)

func New() lifecycle.Handler {
	return func(lc lifecycle.Lifecycle) {
		pidfile.PidPath = filepath.Join(config.GetConfigDir(), pidfile.Name)

		_ = pathutil.IsNotExistMkDir(pidfile.PidPath)

		lc.AfterStart(func(ctx context.Context) error { return pidfile.SavePid() })
	}
}
