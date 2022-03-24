package config

import (
	"path/filepath"

	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config/config_type"
)

var (
	CfgType = "yaml"
	CfgName = "config"
	Home    = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), ".lava")
	CfgPath = ""
	conf    config_type.Config
)

func SetCfg(c config_type.Config) { conf = c }

// GetCfg 获取内存配置
func GetCfg() config_type.Config {
	if conf == nil {
		panic("please init config")
	}

	return conf
}

// Decode decode config to *struct|callback(*struct)
func Decode(name string, fn interface{}) error { return GetCfg().Decode(name, fn) }

// GetMap 通过key获取配置map
func GetMap(keys ...string) config_type.CfgMap { return GetCfg().GetMap(keys...) }
