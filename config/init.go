package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pubgo/dix"
	"github.com/pubgo/golug/env"
	"github.com/pubgo/x/iox"
	"github.com/pubgo/x/osutil"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/x/typex"
	"github.com/pubgo/xerror"
	"github.com/spf13/viper"
)

func On(fn func(cfg *Config)) { xerror.Panic(dix.Dix(fn)) }

func addConfigPath(in string) bool {
	GetCfg().AddConfigPath(in)
	err := GetCfg().ReadInConfig()
	if err == nil {
		return true
	}

	// 检查配置文件是否存在
	if strings.Contains(lower(err.Error()), "not found") {
		return false
	}

	xerror.PanicF(err, "read config failed, path:%s", in)
	return false
}

func initWithCfg() bool {
	if CfgPath == "" {
		return false
	}

	xerror.Assert(pathutil.IsNotExist(CfgPath), "config file not found, path:%s", CfgPath)

	GetCfg().SetConfigFile(CfgPath)

	xerror.PanicF(GetCfg().ReadInConfig(), "config load error, path:%s", CfgPath)

	return true
}

func initWithDir() (err error) {
	defer xerror.RespErr(&err)

	// 指定配置文件
	if initWithCfg() {
		return
	}

	// 检查配置是否存在
	if GetCfg().ReadInConfig() == nil {
		return nil
	}

	home := xerror.PanicStr(osutil.Home())
	var paths = typex.StrOf(
		// 当前${PWD}/config目录
		CfgName,

		// 当前目录${PWD}/home/config目录
		filepath.Join("home", CfgName),

		// etc目录
		filepath.Join("/etc", Domain, Project, CfgName),

		// home工作目录
		filepath.Join(home, ".config", Project, CfgName),
		filepath.Join(home, "."+Domain, Project, CfgName),
	)

	for i := range paths {
		if addConfigPath(paths[i]) {
			return
		}
	}

	return GetCfg().ReadInConfig()
}

// 监控配置中的app自定义配置
func initApp() (err error) {
	defer xerror.RespErr(&err)

	// 处理项目自定义配置
	path := filepath.Dir(GetCfg().ConfigFileUsed())
	appCfg := filepath.Join(path, fmt.Sprint(Project, ".", CfgType))
	if pathutil.IsNotExist(appCfg) {
		return nil
	}

	dt := xerror.PanicStr(iox.ReadText(appCfg))
	// 处理环境变量
	dt = env.Expand(dt)
	// 重新加载配置

	// 合并自定义的配置
	xerror.Panic(GetCfg().MergeConfig(strings.NewReader(dt)))
	return
}

// 处理所有的配置,环境变量和flag
// 配置顺序, 默认值->环境变量->配置文件->flag->配置文件
// 配置文件中可以设置环境变量
// flag可以指定配置文件位置
// 始化配置文件
func Init() (err error) {
	defer xerror.RespErr(&err)

	// 运行环境检查
	var m = RunMode
	switch Mode {
	case m.Dev, m.Stag, m.Prod, m.Test, m.Release:
	default:
		xerror.Assert(true, "running mode does not match, mode: %s", Mode)
	}

	// 配置处理
	cfg = &Config{Viper: viper.New()}

	v := cfg.Viper

	// env 处理
	//v.SetEnvPrefix(EnvPrefix)
	//v.SetEnvKeyReplacer(strings.NewReplacer("_", ".", "-", ".", "/", "."))
	//v.AutomaticEnv()

	// 配置文件名字和类型
	v.SetConfigType(CfgType)
	v.SetConfigName(CfgName)

	// 初始化框架, 加载环境变量, 加载本地配置
	// 初始化完毕所有的配置以及外部配置以及相关的参数和变量
	// 剩下的就是获取配置了
	xerror.PanicF(initWithDir(), "config file load error")
	Home = filepath.Dir(GetCfg().ConfigFileUsed())

	dt := xerror.PanicStr(iox.ReadText(cfg.ConfigFileUsed()))
	// 处理环境变量
	dt = env.Expand(dt)
	// 重新加载配置
	xerror.Panic(cfg.MergeConfig(strings.NewReader(dt)))

	// 加载自定义配置
	xerror.Panic(initApp())

	xerror.Panic(dix.Dix(cfg))
	return nil
}
