package pidfile

import (
	"path/filepath"

	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"
)

func init() {
	di.Provide(func() lifecycle.Handler {
		return func(lc lifecycle.Lifecycle) {
			pidPath = filepath.Join(config.CfgDir, "pidfile")

			_ = pathutil.IsNotExistMkDir(pidPath)

			lc.AfterStart(func() {
				assert.Must(SavePid())
			})
		}
	})
}
