package healthy

import (
	"github.com/pubgo/lug/pkg/ctxutil"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"

	"net/http"
	"time"
)

func init() {
	http.HandleFunc("/health", httpHandle)
}

type health struct {
	Cost string `json:"cost,omitempty"`
	Err  error  `json:"err,omitempty"`
}

func httpHandle(writer http.ResponseWriter, request *http.Request) {
	var dt = make(map[string]*health)
	xerror.Panic(healthList.Each(func(name string, r interface{}) {
		dt[name] = &health{}
		dt[name].Cost = fx.CostWith(func() {
			xerror.TryCatch(func() {
				var ctx, cancel = ctxutil.Timeout(time.Second * 2)
				defer cancel()
				xerror.Panic(r.(HealthCheck)(ctx))
			}, func(err error) {
				dt[name].Err = err
				writer.WriteHeader(http.StatusInternalServerError)
			})
		}).String()
	}))

	var bts, err = jsonx.Marshal(dt)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}

	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(bts)
}
