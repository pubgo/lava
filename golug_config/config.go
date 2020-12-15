package golug_config

import (
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	_ "unsafe"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/valyala/fasttemplate"
)

type Config struct {
	*viper.Viper
}

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("app", pflag.PanicOnError)
	flags.StringVarP(&CfgPath, "cfg", "c", CfgPath, "config path")
	return flags
}

func GetCfg() *Config {
	if cfg == nil {
		xerror.Next().Panic(xerror.New("[config] should be init"))
	}
	return cfg
}

//go:linkname unMarshalReader github.com/spf13/viper.(*Viper).unmarshalReader
func unMarshalReader(v *viper.Viper, in io.Reader, c map[string]interface{}) error

func UnMarshal(path string) map[string]interface{} {
	dt, err := ioutil.ReadFile(path)
	xerror.Next().ExitF(err, path)

	var c = make(map[string]interface{})
	xerror.Next().ExitF(unMarshalReader(GetCfg().Viper, bytes.NewBuffer(dt), c), path)
	return c
}

// Decode
// decode config dataCallback
func Decode(name string, fn interface{}) (err error) {
	defer xerror.RespErr(&err)

	if GetCfg().Get(name) == nil {
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
		ret := reflect.ValueOf(GetCfg().UnmarshalKey).Call(
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
		return xerror.Wrap(GetCfg().UnmarshalKey(name, fn, func(cfg *mapstructure.DecoderConfig) { cfg.TagName = CfgType }))
	default:
		return xerror.Fmt("[fn] type error, type:%#v", fn)
	}

	return
}

func Template(template string) string {
	t := fasttemplate.New(template, "{{", "}}")
	return t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		if strings.HasPrefix(tag, "env_") {
			tag = strings.TrimPrefix(tag, "env_")
			return w.Write([]byte(golug_env.GetEnv(tag)))
		}

		switch tag {
		case "home":
			return w.Write([]byte(golug_env.Home))
		case "trace":
			if golug_env.Trace {
				return w.Write([]byte("true"))
			}
			return w.Write([]byte("false"))
		case "project":
			return w.Write([]byte(golug_env.Project))
		case "domain":
			return w.Write([]byte(golug_env.Domain))
		case "mode":
			return w.Write([]byte(golug_env.Mode))
		case "config":
			return w.Write([]byte(CfgName + "." + CfgType))
		default:
			return w.Write([]byte(GetCfg().GetString(tag)))
		}
	})
}
