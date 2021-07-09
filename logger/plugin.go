package logger

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry/grpc"
	"github.com/pubgo/lug/entry/rest"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/watcher"
	"github.com/pubgo/xerror"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: name,
		OnInit: func(ent interface{}) {
			cfg.Level = runenv.Level
			_ = config.Decode(name, &cfg)

			switch ent := ent.(type) {
			case rest.Entry:
				ent.Use(middleware())
			case grpc.Entry:
				ent.UnaryInterceptor(unaryServer())
				ent.StreamInterceptor(streamServer())
			}

			xerror.Panic(initLog(cfg))
		},
		OnWatch: func(_ string, r *watcher.Response) {
			xerror.Panic(r.Decode(&cfg))
			xerror.Panic(initLog(cfg))
		},
	})
}
