package config

import (
	"path/filepath"
	"strings"

	"github.com/pubgo/xerror"
	"github.com/spf13/viper"
)

var (
	CfgType = "yaml"
	CfgName = "config"
	Home    = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), ".lava")
	CfgPath = ""
	cfg     = &configImpl{v: viper.New()}
)

// Init 初始化配置
func Init() { getCfg().initCfg() }

// GetCfg 获取配置obj
func GetCfg() Config      { return getCfg() }
func getCfg() *configImpl { return cfg }

// Decode decode config to *struct|callback(*struct)
func Decode(name string, fn interface{}) error { return getCfg().Decode(name, fn) }

// GetMap 通过key获取配置map
func GetMap(keys ...string) map[string]interface{} { return getCfg().GetMap(strings.Join(keys, ".")) }
