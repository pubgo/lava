package lifecycle

import "github.com/pubgo/dix/di"

func init() {
	var lc = new(lifecycleImpl)
	di.Provide(func() Handler { return func(lc Lifecycle) {} })
	di.Provide(func() GetLifecycle { return lc })
	di.Provide(func(handlers []Handler) Lifecycle {
		for i := range handlers {
			handlers[i](lc)
		}
		return lc
	})
}
