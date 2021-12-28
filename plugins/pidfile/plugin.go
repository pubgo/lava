package pidfile

import (
	"path/filepath"

	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			pidPath = filepath.Join(config.Home, "pidfile")

			_ = pathutil.IsNotExistMkDir(pidPath)

			p.AfterStart(func() {
				xerror.Panic(SavePid())
			})
		},
	})
}
