package pidfile

import (
	"context"
	"path/filepath"

	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/pathutil"

	"github.com/pubgo/lava/core/lifecycle"
)

func New() lifecycle.Handler {
	return func(lc lifecycle.Lifecycle) {
		pidPath = filepath.Join(config.GetConfigDir(), Name)

		_ = pathutil.IsNotExistMkDir(pidPath)

		lc.AfterStart(func(ctx context.Context) error { return SavePid() })
	}
}
