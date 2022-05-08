package inject

import (
	"github.com/pubgo/xerror"
	"go.uber.org/fx"

	"github.com/pubgo/lava/consts"
)

var factories []fx.Option
var initList []func()

func Register(m fx.Option) {
	xerror.Assert(m == nil, "[m] should not be null")
	factories = append(factories, m)
}

func Name(name string) string {
	if name == consts.KeyDefault {
		name = ""
	}
	return name
}

func Init(fn func()) {
	initList = append(initList, fn)
}

func Load() {
	for i := range initList {
		initList[i]()
	}

	//opts = append(opts, fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
	//	return &fxevent.ZapLogger{Logger: logger.Named("fx")}
	//}))

	xerror.Exit(fx.New(factories...).Err())
}
