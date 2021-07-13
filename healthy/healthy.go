package healthy

import (
	"github.com/pubgo/lug/internal/debug"
	"github.com/pubgo/lug/pkg/ctxutil"
	"github.com/pubgo/lug/types"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/x/try"
	"github.com/pubgo/xerror"

	"net/http"
	"time"
)

func init() {
	debug.On(func(mux *types.DebugMux) {
		mux.Get("/health", httpHandle)
	})
}

type health struct {
	Cost string `json:"cost,omitempty"`
	Err  error  `json:"err,omitempty"`
}

func httpHandle(writer http.ResponseWriter, request *http.Request) {
	var dt = make(map[string]*health)
	healthList.Each(func(name string, r interface{}) {
		dt[name] = &health{}
		dt[name].Cost = fx.CostWith(func() {
			try.Catch(func() {
				var ctx, cancel = ctxutil.Timeout(time.Second * 2)
				defer cancel()
				xerror.Panic(r.(HealthCheck)(ctx))
			}, func(err error) {
				dt[name].Err = err
				writer.WriteHeader(http.StatusInternalServerError)
			})
		}).String()
	})

	var bts, err = jsonx.Marshal(dt)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(bts)
}
