package config

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/pubgo/xerror"
	"github.com/spf13/cast"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/resource"
	"github.com/pubgo/lava/version"
)

var (
	CfgType = "yaml"
	CfgName = "config"
	CfgDir  = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), "."+version.Domain)
	CfgPath = ""
	conf    Config
)

const HomeEnv = "cfg_dir"

func SetCfg(c Config) { conf = c }

// GetCfg 获取内存配置
func GetCfg() Config {
	if conf == nil {
		panic("please init config")
	}

	return conf
}

// Decode decode config to map[string]*struct
func Decode(name string, cfgMap interface{}) {
	defer xerror.RespExit(name)

	xerror.Assert(cfgMap == nil, "[cfgMap] is nil")
	xerror.Assert(reflect.TypeOf(cfgMap).Elem().Kind() != reflect.Map, "[cfgMap] should be map")

	var val = GetCfg().Get(name)
	if val == nil {
		return
	}

	var cfg *typex.RwMap
	for _, data := range cast.ToSlice(val) {
		var dm, err = cast.ToStringMapE(data)
		xerror.Panic(err)

		if cfg == nil {
			cfg = &typex.RwMap{}
		}

		resId := resource.GetResId(dm)

		if _, ok := cfg.Load(resId); ok {
			panic(fmt.Errorf("res=>%s key=>%s,res key already exists", name, resId))
		}

		cfg.Set(resId, dm)
	}

	if cfg == nil {
		cfg = &typex.RwMap{}
		cfg.Set(consts.KeyDefault, val)
	}

	xerror.Panic(merge.Copy(cfgMap, cfg.Map()))
}

// GetMap 通过key获取配置map
func GetMap(keys ...string) CfgMap { return GetCfg().GetMap(keys...) }
