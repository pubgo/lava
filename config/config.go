package config

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/xutil"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/viper"
	"github.com/valyala/fasttemplate"
)

type Config struct {
	*viper.Viper
}

func Map(names ...string) map[string]interface{} {
	return GetCfg().GetStringMap(strings.Join(names, "."))
}

func GetCfg() *Config {
	xerror.Assert(cfg == nil, "[config] please init config")
	return cfg
}

// Decode
// decode config
func Decode(name string, fn interface{}) (b bool) {
	defer xerror.RespExit(name)

	xerror.Assert(name == "" || fn == nil, "[name,fn] should not be nil")
	if GetCfg().Get(name) == nil {
		xlog.Warnf("config key [%s] not found", name)
		return false
	}

	vfn := reflect.ValueOf(fn)
	switch vfn.Type().Kind() {
	case reflect.Func: // func(cfg *Config)
		xerror.Assert(vfn.Type().NumIn() != 1, "[fn] input num should be 1")

		mthIn := reflect.New(vfn.Type().In(0).Elem())
		ret := fx.WrapRaw(GetCfg().UnmarshalKey)(name, mthIn,
			func(cfg *mapstructure.DecoderConfig) { cfg.TagName = "json" })

		if !ret[0].IsNil() {
			xerror.PanicF(ret[0].Interface().(error),
				"config key %s decode error", name)
		}

		vfn.Call(types.ValueOf(mthIn))
	case reflect.Ptr:
		xerror.PanicF(GetCfg().UnmarshalKey(name, fn,
			func(cfg *mapstructure.DecoderConfig) { cfg.TagName = "json" }),
			"config key %s decode error", name)
	default:
		xerror.AssertFn(true, func() string {
			return fmt.Sprintf("[fn] type error, refer: %#v", vfn)
		})
	}

	return true
}

func Template(format string) string {
	t := fasttemplate.New(format, "{{", "}}")
	return t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		tag = strings.TrimSpace(tag)

		// 处理特殊变量
		switch tag {
		case "project_home", "config_home":
			return w.Write(xutil.ToBytes(Home))
		case "trace":
			return w.Write(xutil.ToBytes(strconv.FormatBool(Trace)))
		case "project_name", "project":
			return w.Write(xutil.ToBytes(Project))
		case "domain":
			return w.Write(xutil.ToBytes(Domain))
		case "mode":
			return w.Write(xutil.ToBytes(Mode))
		case "config":
			return w.Write(xutil.ToBytes(CfgName + "." + CfgType))
		case "config_path":
			return w.Write(xutil.ToBytes(GetCfg().ConfigFileUsed()))
		default:
			return w.Write(xutil.ToBytes(GetCfg().GetString(tag)))
		}
	})
}
