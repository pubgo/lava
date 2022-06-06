package debug

import (
	"github.com/pubgo/lava/core/mux"
)

func init() {
	mux.Invoke(func(mux *mux.Mux) {
		mux.Mount("/debug", App())
	})
}
