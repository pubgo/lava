package pidfile

import (
	"path/filepath"

	"github.com/pubgo/dix/di"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/x/pathutil"
)

func init() {
	di.Provide(func() lifecycle.Handler {
		return func(lc lifecycle.Lifecycle) {
			pidPath = filepath.Join(config.CfgDir, "pidfile")

			_ = pathutil.IsNotExistMkDir(pidPath)

			lc.AfterStart(func() {
				SavePid().Must()
			})
		}
	})
}
