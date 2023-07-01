package pidfile

import (
	"path/filepath"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/pathutil"

	"github.com/pubgo/lava/core/lifecycle"
)

func New() lifecycle.Handler {
	return func(lc lifecycle.Lifecycle) {
		pidPath = filepath.Join(config.GetConfigDir(), Name)

		_ = pathutil.IsNotExistMkDir(pidPath)

		lc.AfterStart(func() { assert.Must(SavePid()) })
	}
}
