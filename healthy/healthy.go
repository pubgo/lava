package healthy

import (
	"github.com/pubgo/lug/abc"
	"github.com/pubgo/lug/debug"
	"github.com/pubgo/lug/pkg/ctxutil"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"

	"net/http"
)

func init() {
	debug.On(func(mux *abc.DebugMux) {
		mux.Get("/health", httpHandle)
	})
}

func httpHandle(writer http.ResponseWriter, request *http.Request) {
	var dt = make(map[string]string)
	healthList.Each(func(name string, r interface{}) {
		dt[name] = ""
		if err := r.(HealthCheck)(ctxutil.Default()); err != nil {
			dt[name] = err.Error()
		}
	})

	var bts, err = jsonx.Marshal(dt)
	if err != nil {
		xerror.PanicErr(writer.Write([]byte(err.Error())))
		return
	}

	xerror.PanicErr(writer.Write(bts))
}
