package golug_config

import (
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/viper"
)

func Init() (err error) {
	defer xerror.RespErr(&err)

	// 从环境变量中获取系统默认值
	// 获取系统默认的前缀, 环境变量前缀等
	golug_env.Get(&Domain, "golug", "golug_domain", "golug_prefix", "env_prefix")
	if Domain = strings.TrimSpace(strings.ToLower(Domain)); Domain == "" {
		Domain = "golug"
		xlog.Warnf("[domain] prefix should be set, default: %s", Domain)
	}

	// 设置系统环境变量前缀
	golug_env.Prefix(Domain)

	// 使用前缀获取系统环境变量
	golug_env.Get(&Home, "home", "dir")
	golug_env.Get(&Project, "project", "name")
	golug_env.Get(&Mode, "mode", "run")

	CfgPath = filepath.Join(Home, "config", "config.yaml")
	if !golug_util.PathExist(Home) {
		xerror.Panic(xerror.Fmt("home path [%s] not exists", Home))
	}
	if !golug_util.PathExist(CfgPath) {
		xerror.Panic(xerror.Fmt("config path [%s] not exists", CfgPath))
	}

	{
		// 配置viper
		initViperEnv(Domain)

		// 配置文件名字和类型
		viper.SetConfigType(CfgType)
		viper.SetConfigName(CfgName)

		// 监控默认配置
		viper.AddConfigPath(Home)
		viper.AddConfigPath(filepath.Join(Home, CfgName))

		// 监控当前工作目录
		_pwd := xerror.PanicStr(filepath.Abs(filepath.Dir("")))
		viper.AddConfigPath(_pwd)
		viper.AddConfigPath(filepath.Join(_pwd, CfgName))

		// 监控Home工作目录
		_home := xerror.PanicErr(homedir.Dir()).(string)
		viper.AddConfigPath(filepath.Join(_home, "."+Project))
		viper.AddConfigPath(filepath.Join(_home, "."+Project, CfgName))

		// 检查配置文件是否存在
		if err := viper.ReadInConfig(); err != nil && !strings.Contains(err.Error(), "not found") {
			xerror.ExitF(err, "read config failed")
		}

		// 获取配置文件所在目录
		Home = filepath.Dir(xerror.PanicStr(filepath.Abs(viper.ConfigFileUsed())))
	}

	//_, err = cfg.Load("watcher")
	//if err != nil {
	//	xlog.Debugf("config [watcher] is error: %v", err)
	//}

	return nil
}
