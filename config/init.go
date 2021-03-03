package config

import (
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pubgo/dix"
	"github.com/pubgo/golug/env"
	"github.com/pubgo/golug/golug"
	"github.com/pubgo/golug/gutils"
	"github.com/pubgo/xerror"
	"github.com/spf13/viper"
)

// 默认的全局配置
var (
	name    = "config"
	CfgType = "yaml"
	CfgName = "config"
	cfg     *Config
)

func init() {
	env.GetVal(&CfgType, "cfg_type")
	env.GetVal(&CfgName, "cfg_name")
}

func On(fn func(cfg *Config)) { xerror.Panic(dix.Dix(fn)) }

func initWithDir() (err error) {
	defer xerror.RespErr(&err)

	v := GetCfg()

	// config 路径
	// 当前目录home目录
	v.AddConfigPath(filepath.Join("home", CfgName))

	// 检查配置文件是否存在
	if err := v.ReadInConfig(); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "not found") {
			return xerror.WrapF(err, "read config failed")
		}
	}

	// etc目录
	v.AddConfigPath(filepath.Join("/etc", golug.Domain, golug.Project, CfgName))

	// 监控Home工作目录
	home := xerror.PanicStr(homedir.Dir())
	v.AddConfigPath(filepath.Join(home, ".config", golug.Project, CfgName))
	v.AddConfigPath(filepath.Join(home, "."+golug.Domain, golug.Project, CfgName))

	if err := v.ReadInConfig(); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "not found") {
			return xerror.WrapF(err, "read config failed")
		}
	}
	return nil
}

// 监控配置中的其他配置文件
func initApp() (err error) {
	defer xerror.RespErr(&err)

	v := GetCfg()

	// 处理独立的组件的配置, config.nsq.yaml, config.mysql.yaml
	appCfg := filepath.Join(filepath.Dir(v.ConfigFileUsed()), "app."+CfgType)
	if !gutils.PathExist(appCfg) {
		return nil
	}

	// 从自定义文件中解析配置
	val1 := UnMarshal(appCfg)
	if val1 == nil {
		return
	}

	// 合并自定义的配置
	for key, val2 := range val1 {
		// 获取config中默认的配置
		if val := v.GetStringMap(key); val != nil {
			// 合并配置
			gutils.Mergo(&val, val2)
			val2 = val
		}
		v.Set(key, val2)
	}
	return nil
}

// 处理所有的配置,环境变量和flag
// 配置顺序, 默认值->环境变量->配置文件->flag->配置文件
// 配置文件中可以设置环境变量
// flag可以指定配置文件位置
// 始化配置文件
func Init() (err error) {
	defer xerror.RespErr(&err)

	// 运行环境检查
	var m = golug.RunMode
	switch golug.Mode {
	case m.Dev, m.Stag, m.Prod, m.Test, m.Release:
	default:
		xerror.Assert(true, "running mode does not match, mode: %s", golug.Mode)
	}

	// 配置处理
	cfg = &Config{Viper: viper.New()}

	v := cfg.Viper

	// env 处理
	v.SetEnvPrefix(golug.Domain)
	v.SetEnvKeyReplacer(strings.NewReplacer("_", ".", "-", ".", "/", "."))
	v.AutomaticEnv()

	// 把环境变量的值设置到全局配置当中
	for key, val := range env.List() {
		v.Set(key, val)
	}

	// 配置文件名字和类型
	v.SetConfigType(CfgType)
	v.SetConfigName(CfgName)

	// 初始化框架, 加载环境变量, 加载本地配置
	// 初始化完毕所有的配置以及外部配置以及相关的参数和变量
	// 剩下的就是获取配置了
	if cfg.ReadInConfig() != nil {
		xerror.Panic(initWithDir())
	}

	xerror.Assert(cfg.ConfigFileUsed() == "", "config file not found")
	xerror.ExitF(cfg.ReadInConfig(), "read config failed")
	xerror.Panic(initApp())
	xerror.Panic(dix.Dix(cfg))
	return nil
}
