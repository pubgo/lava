package cmux

import (
	"errors"
	"net"
	"time"

	"github.com/pubgo/dix"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
)

func init() {
	dix.Register(func(c *config.App, log *logging.Logger) *Mux {
		return &Mux{
			Addr:        c.Addr,
			ReadTimeout: time.Second * 2,
			HandleError: func(err error) bool {
				if errors.Is(err, net.ErrClosed) {
					return true
				}

				log.Named("cmux").Error("cmux matcher failed", logutil.ErrField(err)...)
				return true
			},
		}
	})
}
