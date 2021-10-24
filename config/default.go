package config

import (
	"path/filepath"
	"strings"

	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/pubgo/lava/pkg/env"
)

var (
	CfgType = "yaml"
	CfgName = "config"
	Home    = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), ".lava")
	CfgPath = ""
	cfg     = &configImpl{v: viper.New()}
)

func init() {
	env.GetWith(&CfgType, "cfg_type", "config_type")
	env.GetWith(&CfgName, "cfg_name", "config_name")
	env.GetWith(&Home, "project_home", "config_home", "config_dir")
}

func Init() error                                  { return getCfg().Init() }
func GetCfg() Config                               { return getCfg() }
func Decode(name string, fn interface{}) error     { return getCfg().Decode(name, fn) }
func GetMap(keys ...string) map[string]interface{} { return getCfg().GetMap(strings.Join(keys, ".")) }

func getCfg() *configImpl {
	xerror.Assert(cfg == nil, "[config] please init config")
	return cfg
}

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("", pflag.PanicOnError)
	flags.StringVarP(&CfgPath, "config", "c", CfgPath, "config path")
	return flags
}
