package broker

import (
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
)

func init() {
	vars.Watch(Name+"_factory", func() interface{} {
		var data = make(map[string]string)
		xerror.Panic(factories.Each(func(name string, fc Factory) {
			data[name] = stack.Func(fc)
		}))
		return data
	})

	vars.Watch(Name+"_broker", func() interface{} {
		var data = make(map[string]string)
		xerror.Panic(brokers.Each(func(name string, fc Broker) {
			data[name] = fc.String()
		}))
		return data
	})
}
