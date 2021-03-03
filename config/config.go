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
	"github.com/pubgo/golug/types"
	"github.com/pubgo/x/xutil"
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
// UnMarshal config from file to map
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
	defer xerror.RespExit(name)

	xerror.Assert(fn == nil, "[fn] should not be nil")
	xerror.Assert(GetCfg().Get(name) == nil, "[name] config not found")

	vfn := reflect.ValueOf(fn)
	switch vfn.Type().Kind() {
	case reflect.Func: // func(cfg *Config)
		xerror.Assert(vfn.Type().NumIn() != 1, "[fn] input num should be one")

		mthIn := reflect.New(vfn.Type().In(0).Elem())
		ret := reflect.ValueOf(GetCfg().UnmarshalKey).
			Call(types.ValueOf(
				reflect.ValueOf(name), mthIn,
				reflect.ValueOf(func(cfg *mapstructure.DecoderConfig) { cfg.TagName = "json" }),
			))
		if !ret[0].IsNil() {
			xerror.PanicF(ret[0].Interface().(error), "%s config decode error", name)
		}

		vfn.Call(types.ValueOf(mthIn))
	case reflect.Ptr:
		xerror.Panic(GetCfg().UnmarshalKey(name, fn,
			func(cfg *mapstructure.DecoderConfig) { cfg.TagName = "json" }))
	default:
		xerror.Assert(true,"[fn] type error, type:%#v", vfn)
	}
}

func Template(format string) string {
	t := fasttemplate.New(format, "{{", "}}")
	return t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		tag = strings.TrimSpace(tag)

		// 处理特殊变量
		switch tag {
		case "home":
			return w.Write(xutil.ToBytes(golug.Home))
		case "trace":
			return w.Write(xutil.ToBytes(strconv.FormatBool(golug.Trace)))
		case "project":
			return w.Write(xutil.ToBytes(golug.Project))
		case "domain":
			return w.Write(xutil.ToBytes(golug.Domain))
		case "mode":
			return w.Write(xutil.ToBytes(golug.Mode))
		case "config":
			return w.Write(xutil.ToBytes(CfgName + "." + CfgType))
		default:
			return w.Write(xutil.ToBytes(GetCfg().GetString(tag)))
		}
	})
}
