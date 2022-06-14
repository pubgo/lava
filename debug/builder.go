package debug

import (
	"github.com/pubgo/lava/core/router"
)

func init() {
	router.Register("/debug", App())
}
