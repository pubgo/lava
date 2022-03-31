package config

import (
	"path/filepath"

	"github.com/pubgo/xerror"

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

// Decode decode config to *struct|callback(*struct)
func Decode(name string, fn interface{}) error { return GetCfg().Decode(name, fn) }

// GetMap 通过key获取配置map
func GetMap(keys ...string) CfgMap { return GetCfg().GetMap(keys...) }
