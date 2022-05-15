package inject

import (
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/xerror"
	"go.uber.org/fx"
)

var options []fx.Option

func Annotated(aa fx.Annotated) {
	Register(fx.Provide(aa))
}

func RegisterName(name string, target interface{}) {
	Register(fx.Provide(fx.Annotated{
		Name:   Name(name),
		Target: target,
	}))
}

func RegisterGroup(name string, target interface{}) {
	Register(fx.Provide(fx.Annotated{
		Group:  Name(name),
		Target: target,
	}))
}

func NameGroup(group string, name string, target interface{}) {
	Register(fx.Provide(fx.Annotated{
		Name:   Name(name),
		Target: target,
	}))

	Register(fx.Provide(fx.Annotated{
		Group:  group,
		Target: target,
	}))
}

func Provide(c ...interface{}) {
	Register(fx.Provide(c...))
}

func Invoke(funcs ...interface{}) {
	Register(fx.Invoke(funcs...))
}

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

func Load() {
	//opts = append(opts, fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
	//	return &fxevent.ZapLogger{Logger: logger.Named("fx")}
	//}))

	//options = append(options, fx.WithLogger(
	//	func() fxevent.Logger {
	//		return fxevent.NopLogger
	//	},
	//))
	xerror.Exit(fx.New(options...).Err())
}
