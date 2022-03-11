package config

import (
	"path/filepath"

	"github.com/pubgo/xerror"
	"github.com/spf13/viper"

	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/watcher"
)

var (
	CfgType = "yaml"
	CfgName = "config"
	Home    = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), ".lava")
	CfgPath = ""
	cfg     = &configImpl{v: viper.New()}
)

// Init 配置初始化
func Init() {
	cfg.initCfg()

	watcher.Init(getCfg())
}

// GetCfg 获取内存配置
func GetCfg() config_type.Interface { return getCfg() }
func getCfg() *configImpl {
	if !cfg.init {
		panic("please init config")
	}

	return cfg
}

// Decode decode config to *struct|callback(*struct)
func Decode(name string, fn interface{}) error { return getCfg().Decode(name, fn) }

// GetMap 通过key获取配置map
func GetMap(keys ...string) types.CfgMap { return getCfg().GetMap(keys...) }
