package config

import (
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	_ "unsafe"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/golug/golug"
	"github.com/pubgo/xerror"
	"github.com/spf13/viper"
	"github.com/valyala/fasttemplate"
)

type Config struct {
	*viper.Viper
}

func GetCfg() *Config {
	xerror.Assert(cfg == nil, "[config] please init config")
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
	defer xerror.RespRaise(func(err xerror.XErr) error { return xerror.WrapF(err, "name:%s", name) })

	if GetCfg().Get(name) == nil {
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
		tag = strings.TrimSpace(tag)

		// 处理特殊变量
		switch tag {
		case "home":
			return w.Write([]byte(golug.Home))
		case "trace":
			return w.Write([]byte(strconv.FormatBool(golug.Trace)))
		case "project":
			return w.Write([]byte(golug.Project))
		case "domain":
			return w.Write([]byte(golug.Domain))
		case "mode":
			return w.Write([]byte(golug.Mode))
		case "config":
			return w.Write([]byte(CfgName + "." + CfgType))
		default:
			return w.Write([]byte(GetCfg().GetString(tag)))
		}
	})
}