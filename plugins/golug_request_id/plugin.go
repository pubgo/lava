package golug_request_id

import (
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
)

func init() {
	xerror.Exit(golug_plugin.Register(&golug_plugin.Base{
		Enabled: true,
		Name:    name,
		OnInit: func(ent golug_entry.Entry) {
			xerror.Panic(ent.UnWrap(func(entry golug_entry.HttpEntry) { entry.Use(httpRequestId()) }))
			xerror.Panic(ent.UnWrap(func(entry golug_entry.GrpcEntry) {
				entry.UnaryServer(grpcUnaryServer())
				entry.StreamServer(grpcStreamServer())
			}))
		},
	}))
}
