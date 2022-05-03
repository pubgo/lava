package pidfile

import (
	"path/filepath"

	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			pidPath = filepath.Join(config.CfgDir, "pidfile")

			_ = pathutil.IsNotExistMkDir(pidPath)

			p.AfterStart(func() {
				xerror.Panic(SavePid())
			})
		},
	})
}
