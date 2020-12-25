package golug_config

import (
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	_ "unsafe"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/viper"
	"github.com/valyala/fasttemplate"
)

type Config struct {
	*viper.Viper
}

func GetCfg() *Config {
	if cfg == nil {
		xerror.Panic(xerror.New("[config] should be init"))
	}
	return cfg
}

//go:linkname unMarshalReader github.com/spf13/viper.(*Viper).unmarshalReader
func unMarshalReader(v *viper.Viper, in io.Reader, c map[string]interface{}) error

// UnMarshal
// UnMarshal config to map
func UnMarshal(path string) map[string]interface{} {
	dt, err := ioutil.ReadFile(path)
	xerror.ExitF(err, path)

	var c = make(map[string]interface{})
	xerror.ExitF(unMarshalReader(GetCfg().Viper, bytes.NewBuffer(dt), c), path)
	return c
}

// Decode
// decode config
func Decode(name string, fn interface{}) {
	defer xerror.RespRaise("name:%s", name)

	if GetCfg().Get(name) == nil {
		xlog.Warnf("%s not found", name)
		return
	}

	if fn == nil {
		xerror.Panic(xerror.New("[fn] should not be nil"))
	}

	vfn := reflect.ValueOf(fn)
	switch vfn.Type().Kind() {
	case reflect.Func: // func(cfg *Config)
		if vfn.Type().NumIn() != 1 {
			xerror.Panic(xerror.New("[fn] input num should be one"))
		}

		mthIn := reflect.New(vfn.Type().In(0).Elem())
		ret := reflect.ValueOf(GetCfg().UnmarshalKey).Call(
			[]reflect.Value{
				reflect.ValueOf(name), mthIn,
				reflect.ValueOf(func(cfg *mapstructure.DecoderConfig) { cfg.TagName = "json" }),
			},
		)
		if !ret[0].IsNil() {
			xerror.PanicF(ret[0].Interface().(error), "%s config decode error", name)
		}

		vfn.Call([]reflect.Value{mthIn})
	case reflect.Ptr:
		xerror.Panic(GetCfg().UnmarshalKey(name, fn, func(cfg *mapstructure.DecoderConfig) { cfg.TagName = "json" }))
	default:
		xerror.Panic(xerror.Fmt("[fn] type error, type:%#v", fn))
	}
}

func Template(format string) string {
	t := fasttemplate.New(format, "{{", "}}")
	return t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		tag = trim(tag)

		// 处理环境变量, env_前缀的为环境变量
		if strings.HasPrefix(tag, "env_") {
			tag = strings.TrimPrefix(tag, "env_")
			return w.Write([]byte(golug_env.GetSysEnv(golug_env.WithPrefix(tag), golug_env.TrimPrefix(tag))))
		}

		// 处理特殊变量
		switch tag {
		case "home":
			return w.Write([]byte(golug_app.Home))
		case "trace":
			if golug_app.Trace {
				return w.Write([]byte("true"))
			}
			return w.Write([]byte("false"))
		case "project":
			return w.Write([]byte(golug_app.Project))
		case "domain":
			return w.Write([]byte(golug_app.Domain))
		case "mode":
			return w.Write([]byte(golug_app.Mode))
		case "config":
			return w.Write([]byte(CfgName + "." + CfgType))
		default:
			return w.Write([]byte(GetCfg().GetString(tag)))
		}
	})
}
