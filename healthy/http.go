package healthy

import (
	"net/http"
	"time"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/ctxutil"
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
			xerror.TryCatch(func() (interface{}, error) {
				return nil, r.(HealthCheck)(ctxutil.Timeout(time.Second * 2))
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
