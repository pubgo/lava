package golug_broker_nsq

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/golug_broker"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/plugins/golug_nsq"
	"github.com/pubgo/xlog"
)

func init() {
	if golug_env.Trace {
		xlog.Debug("init ok")
	}
}

func init() {
	golug.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
		for k, v := range golug_broker.GetCfg() {
			if v.Driver != golug_nsq.Name {
				continue
			}

			_, ok := golug_nsq.GetCfg()[v.Name]
			if !ok {
				continue
			}

			golug_broker.Register(k, &nsqBroker{name: v.Name})
		}
	})
}
