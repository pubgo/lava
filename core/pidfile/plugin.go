package pidfile

import (
	"path/filepath"

	"github.com/pubgo/dix"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/lifecycle"
)

func init() {
	dix.Provider(func() lifecycle.Handler {
		return func(lc lifecycle.Lifecycle) {
			pidPath = filepath.Join(config.CfgDir, "pidfile")

			_ = pathutil.IsNotExistMkDir(pidPath)

			lc.AfterStarts(func() {
				xerror.Panic(SavePid())
			})
		}
	})
}
