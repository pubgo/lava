package pidfile

import (
	"path/filepath"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/lifecycle"
	"github.com/pubgo/funk/pathutil"
)

func New() lifecycle.Handler {
	return func(lc lifecycle.Lifecycle) {
		pidPath = filepath.Join(config.CfgDir, "pidfile")

		_ = pathutil.IsNotExistMkDir(pidPath)

		lc.AfterStart(func() {
			assert.Must(SavePid())
		})
	}
}
