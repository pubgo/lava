package config

import (
	"bytes"
	"github.com/pubgo/xlog"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	_ "unsafe"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/golug/types"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/xutil"
	"github.com/pubgo/xerror"
	"github.com/spf13/viper"
	"github.com/valyala/fasttemplate"
)

type Config struct {
	*viper.Viper
}

func Map(names ...string) map[string]interface{} {
	return GetCfg().GetStringMap(strings.Join(names, "."))
}

func Enabled(name string) bool { return GetCfg().GetBool("app." + name) }

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

	xerror.Assert(name == "" || fn == nil, "[name,fn] should not be nil")
	if GetCfg().Get(name) == nil{
		xlog.Debugf("config key [%s] not found",name)
		return
	}

	vfn := reflect.ValueOf(fn)
	switch vfn.Type().Kind() {
	case reflect.Func: // func(cfg *Config)
		xerror.Assert(vfn.Type().NumIn() != 1, "[fn] input num should be one")

		mthIn := reflect.New(vfn.Type().In(0).Elem())
		ret := fx.WrapRaw(GetCfg().UnmarshalKey)(name, mthIn,
			func(cfg *mapstructure.DecoderConfig) { cfg.TagName = "json" })

		if !ret[0].IsNil() {
			xerror.PanicF(ret[0].Interface().(error), "%s config decode error", name)
		}

		vfn.Call(types.ValueOf(mthIn))
	case reflect.Ptr:
		xerror.Panic(GetCfg().UnmarshalKey(name, fn,
			func(cfg *mapstructure.DecoderConfig) { cfg.TagName = "json" }))
	default:
		xerror.Assert(true, "[fn] type error, type:%#v", vfn)
	}
}

func Template(format string) string {
	t := fasttemplate.New(format, "{{", "}}")
	return t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		tag = strings.TrimSpace(tag)

		// 处理特殊变量
		switch tag {
		case "home":
			return w.Write(xutil.ToBytes(Home))
		case "trace":
			return w.Write(xutil.ToBytes(strconv.FormatBool(Trace)))
		case "project":
			return w.Write(xutil.ToBytes(Project))
		case "domain":
			return w.Write(xutil.ToBytes(Domain))
		case "mode":
			return w.Write(xutil.ToBytes(Mode))
		case "config":
			return w.Write(xutil.ToBytes(CfgName + "." + CfgType))
		default:
			return w.Write(xutil.ToBytes(GetCfg().GetString(tag)))
		}
	})
}
