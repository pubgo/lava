package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pubgo/dix"
	"github.com/pubgo/lug/app"
	"github.com/pubgo/lug/pkg/env"
	"github.com/pubgo/x/iox"
	"github.com/pubgo/x/osutil"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/x/typex"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	env.GetWith(&CfgType, "cfg_type", "config_type")
	env.GetWith(&CfgName, "cfg_name", "config_name")
	env.GetWith(&Home, "project_home", "config_home", "cfg_dir", "config_path", "config_dir")
}

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("app", pflag.PanicOnError)
	flags.StringVarP(&CfgPath, "cfg", "c", CfgPath, "config path")
	return flags
}

func On(fn func(cfg Config)) { xerror.Panic(dix.Dix(fn)) }

func addConfigPath(v *viper.Viper, in string) bool {
	v.AddConfigPath(in)
	err := v.ReadInConfig()
	if err == nil {
		return true
	}

	// 检查配置文件是否存在
	if strings.Contains(strings.ToLower(err.Error()), "not found") {
		return false
	}

	xerror.PanicF(err, "read config failed, path:%s", in)
	return false
}

func initWithCfg(v *viper.Viper) bool {
	if CfgPath == "" {
		return false
	}

	xerror.Assert(pathutil.IsNotExist(CfgPath), "config file not found, path:%s", CfgPath)

	v.SetConfigFile(CfgPath)

	xerror.PanicF(v.ReadInConfig(), "config load error, path:%s", CfgPath)

	return true
}

func initWithDir(v *viper.Viper) (err error) {
	defer xerror.RespErr(&err)

	// 指定配置文件
	if initWithCfg(v) {
		return
	}

	// 检查配置是否存在
	if v.ReadInConfig() == nil {
		return nil
	}

	home := xerror.PanicStr(osutil.Home())
	var paths = typex.StrOf(
		// 当前${PWD}/config目录
		CfgName,

		// 当前目录${PWD}/home/config目录
		filepath.Join("home", CfgName),

		// etc目录
		filepath.Join("/etc", app.Domain, app.Project, CfgName),

		// home工作目录
		filepath.Join(home, ".config", app.Project, CfgName),
		filepath.Join(home, "."+app.Domain, app.Project, CfgName),
	)

	for i := range paths {
		if addConfigPath(v, paths[i]) {
			return
		}
	}

	return v.ReadInConfig()
}

// 监控配置中的app自定义配置
func initApp(v *viper.Viper) (err error) {
	defer xerror.RespErr(&err)

	// 处理项目自定义配置
	path := filepath.Dir(v.ConfigFileUsed())
	appCfg := filepath.Join(path, fmt.Sprint(app.Project, ".", CfgType))
	if pathutil.IsNotExist(appCfg) {
		return nil
	}

	dt := xerror.PanicStr(iox.ReadText(appCfg))

	// 处理环境变量
	dt = env.Expand(dt)

	// 合并自定义的配置
	xerror.Panic(v.MergeConfig(strings.NewReader(dt)))
	return
}

// Init 处理所有的配置,环境变量和flag
// 配置顺序, 默认值->环境变量->配置文件->flag
// 配置文件中可以设置环境变量
// flag可以指定配置文件位置
// 始化配置文件
func (t *conf) Init() (err error) {
	defer xerror.RespErr(&err)

	t.rw.Lock()
	defer t.rw.Unlock()

	xerror.Assert(!app.CheckMode(), "")

	// 配置处理
	v := t.v

	// env 处理
	//v.SetEnvPrefix(EnvPrefix)
	//v.SetEnvKeyReplacer(strings.NewReplacer("_", ".", "-", ".", "/", "."))
	//v.AutomaticEnv()

	// 配置文件名字和类型
	v.SetConfigType(CfgType)
	v.SetConfigName(CfgName)

	// 初始化框架, 加载环境变量, 加载本地配置
	// 初始化完毕所有的配置以及外部配置以及相关的参数和变量
	// 然后获取配置了
	xerror.PanicF(initWithDir(v), "config file load error")
	Home = filepath.Dir(filepath.Dir(v.ConfigFileUsed()))

	dt := xerror.PanicStr(iox.ReadText(v.ConfigFileUsed()))
	// 处理环境变量
	dt = env.Expand(dt)
	// 重新加载配置
	xerror.Panic(v.MergeConfig(strings.NewReader(dt)))

	// 加载自定义配置
	xerror.Panic(initApp(v))
	return nil
}
