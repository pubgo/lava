package app

import (
	"errors"
	"net"
	"time"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/cmux"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
)

const Name = "app"

func init() {
	dix.Register(func(c config.Config, log *logging.Logger) *cmux.Mux {
		var cfg Cfg
		xerror.Panic(c.UnmarshalKey(Name, &cfg))
		xerror.Panic(cfg.Check())

		return &cmux.Mux{
			Addr:        cfg.Addr,
			ReadTimeout: time.Second * 2,
			HandleError: func(err error) bool {
				if errors.Is(err, net.ErrClosed) {
					return true
				}

				log.Named(Name).Error("cmux matcher failed", logutil.ErrField(err)...)
				return true
			},
		}
	})
}
