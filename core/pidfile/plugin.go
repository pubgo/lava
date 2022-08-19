package pidfile

import (
	"path/filepath"

	"github.com/pubgo/dix"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/x/pathutil"
)

func init() {
	dix.Provider(func() lifecycle.Handler {
		return func(lc lifecycle.Lifecycle) {
			pidPath = filepath.Join(config.CfgDir, "pidfile")

			_ = pathutil.IsNotExistMkDir(pidPath)

			lc.AfterStart(SavePid)
		}
	})
}
