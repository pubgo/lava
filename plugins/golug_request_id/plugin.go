package golug_request_id

import (
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/golug_grpc"
	"github.com/pubgo/golug/golug_entry/golug_rest"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
)

func init() {
	xerror.Exit(golug_plugin.Register(&golug_plugin.Base{
		Name: name,
		OnInit: func(ent golug_entry.Entry) {
			ent.UnWrap(func(entry golug_rest.Entry) { entry.Use(httpRequestId()) })
			ent.UnWrap(func(entry golug_grpc.Entry) {
				entry.UnaryServer(grpcUnaryServer())
				entry.StreamServer(grpcStreamServer())
			})
		},
	}))
}
