package golug_config

import (
	"os"
	"reflect"
	"syscall"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// 默认的全局配置
var (
	CfgType = "yaml"
	CfgName = "config"
	Debug   = true
	IsBlock = true
	cfg     *Config
	Signal  os.Signal = syscall.Signal(0)
)

type Config struct {
	*viper.Viper
}

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("app", pflag.PanicOnError)
	flags.StringVarP(&golug_env.Mode, "mode", "m", golug_env.Mode, "running mode(dev|test|stag|prod|release)")
	flags.StringVarP(&golug_env.Home, "home", "c", golug_env.Home, "project config home dir")
	flags.BoolVarP(&Debug, "debug", "d", Debug, "enable debug")
	flags.BoolVarP(&golug_env.Trace, "trace", "t", golug_env.Trace, "enable trace")
	flags.BoolVarP(&IsBlock, "block", "b", IsBlock, "enable signal block")
	return flags
}

func GetCfg() *Config {
	if cfg == nil {
		xerror.Panic(xerror.New("config should be init"))
	}
	return cfg
}

// Decode
// decode config data
func Decode(name string, fn interface{}) (err error) {
	defer xerror.RespErr(&err)

	if viper.Get(name) == nil {
		return nil
	}

	if fn == nil {
		return xerror.New("fn should not be nil")
	}

	vfn := reflect.ValueOf(fn)
	switch vfn.Type().Kind() {
	case reflect.Func:
		if vfn.Type().NumIn() != 1 {
			return xerror.New("[fn] input num should be one")
		}

		mthIn := reflect.New(vfn.Type().In(0).Elem())
		ret := reflect.ValueOf(viper.UnmarshalKey).Call(
			[]reflect.Value{
				reflect.ValueOf(name), mthIn,
				reflect.ValueOf(func(cfg *mapstructure.DecoderConfig) { cfg.TagName = CfgType }),
			},
		)
		if !ret[0].IsNil() {
			return xerror.WrapF(ret[0].Interface().(error), "config decode error")
		}

		vfn.Call([]reflect.Value{mthIn})
	case reflect.Ptr:
		return xerror.Wrap(viper.UnmarshalKey(name, fn, func(cfg *mapstructure.DecoderConfig) { cfg.TagName = CfgType }))
	default:
		return xerror.Fmt("[fn] type error, type:%#v", fn)
	}

	return
}
