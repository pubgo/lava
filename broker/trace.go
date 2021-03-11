package broker

import (
	"github.com/pubgo/golug/tracelog"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
)

func init() {
	tracelog.Watch(Name+"_factory", func() interface{} {
		var data = make(map[string]string)
		xerror.Panic(factories.Each(func(name string, fc Factory) {
			data[name] = stack.Func(fc)
		}))
		return data
	})

	tracelog.Watch(Name+"_broker", func() interface{} {
		var data = make(map[string]string)
		xerror.Panic(brokers.Each(func(name string, fc Broker) {
			data[name] = fc.String()
		}))
		return data
	})
}
