package healthy

import (
	"net/http"

	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/mux"
	"github.com/pubgo/lava/pkg/utils"
	"github.com/pubgo/lava/types"
)

func init() {
	mux.Get("/health", httpHandle)
}

type health struct {
	Cost string `json:"cost,omitempty"`
	Err  error  `json:"err,omitempty"`
	Msg  string `json:"err_msg,omitempty"`
}

func httpHandle(writer http.ResponseWriter, request *http.Request) {
	var dt = make(map[string]*health)
	xerror.Panic(healthList.Each(func(name string, r interface{}) {
		var h = &health{}
		var dur, err = utils.Cost(func() { xerror.Panic(r.(types.Healthy)(request)) })
		h.Cost = dur.String()
		if err != nil {
			h.Msg = err.Error()
			h.Err = err
		}
		dt[name] = h
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
