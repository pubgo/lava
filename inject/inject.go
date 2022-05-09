package inject

import (
	"github.com/pubgo/xerror"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/pubgo/lava/consts"
)

var options []fx.Option
var initList []func()

func Register(m fx.Option) {
	xerror.Assert(m == nil, "[m] should not be null")
	options = append(options, m)
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

	options = append(options, fx.WithLogger(
		func() fxevent.Logger {
			return fxevent.NopLogger
		},
	))

	xerror.Exit(fx.New(options...).Err())
}
