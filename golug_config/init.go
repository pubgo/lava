package golug_config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/imdario/mergo"
	"github.com/mitchellh/go-homedir"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/golug_plugin/plugins/golug_etcd"
	"github.com/pubgo/golug/golug_watcher"
	"github.com/pubgo/golug/golug_watcher/watchers/etcd"
	"github.com/pubgo/golug/golug_watcher/watchers/file"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/clientv3"
)

func Init() (err error) {
	defer xerror.RespErr(&err)

	// 从环境变量中获取系统默认值
	// 获取系统默认的前缀, 环境变量前缀等
	golug_env.Get(&golug_env.Domain, "golug", "golug_domain", "golug_prefix", "env_prefix")
	if golug_env.Domain = strings.TrimSpace(strings.ToLower(golug_env.Domain)); golug_env.Domain == "" {
		golug_env.Domain = "golug"
		xlog.Warnf("[domain] prefix should be set, default: %s", golug_env.Domain)
	}
	golug_env.Get(&golug_env.Project, "project", "name", "server_name")

	{
		cfg = &Config{Viper: viper.New()}
		v := cfg.Viper

		if CfgPath != "" {
			CfgPath = xerror.PanicStr(filepath.Abs(CfgPath))
			CfgPath = xerror.PanicStr(filepath.EvalSymlinks(CfgPath))
			CfgType = filepath.Ext(CfgPath)
			CfgName = strings.TrimRight(filepath.Base(CfgPath), "."+CfgType)
			v.SetConfigFile(CfgPath)
			golug_env.Home = filepath.Dir(filepath.Dir(CfgPath))
		}

		// 配置文件名字和类型
		v.SetConfigType(CfgType)
		v.SetConfigName(CfgName)

		// config 路径
		v.AddConfigPath(".")
		v.AddConfigPath(fmt.Sprintf("/etc/%s/%s/", golug_env.Domain, golug_env.Project))
		v.AddConfigPath(fmt.Sprintf("$HOME/.%s/%s", golug_env.Domain, golug_env.Project))
		v.SetEnvPrefix(golug_env.Domain)
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "/"))
		v.AutomaticEnv()

		// 监控默认配置
		v.AddConfigPath(filepath.Join(golug_env.Home, CfgName))

		// 监控当前工作目录
		_pwd := xerror.PanicStr(filepath.Abs(filepath.Dir("")))
		v.AddConfigPath(filepath.Join(_pwd, CfgName))

		// 监控Home工作目录
		_home := xerror.PanicErr(homedir.Dir()).(string)
		v.AddConfigPath(filepath.Join(_home, "."+golug_env.Project, CfgName))

		// 检查配置文件是否存在
		xerror.PanicF(v.ReadInConfig(), "read config failed")

		// 获取配置文件所在目录
		golug_env.Home = filepath.Dir(filepath.Dir(xerror.PanicStr(filepath.Abs(v.ConfigFileUsed()))))

		xerror.Exit(filepath.Walk(filepath.Join(golug_env.Home, "config"), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return xerror.Wrap(err)
			}

			if info.IsDir() {
				return nil
			}

			// 配置文件类型检查
			if !strings.HasSuffix(info.Name(), CfgType) {
				return nil
			}

			// 文件名字检查
			if info.Name() == CfgName+"."+CfgType {
				return nil
			}

			ns := strings.Split(info.Name(), ".")
			if len(ns) != 3 {
				xerror.Exit(xerror.Fmt("config name error, %s", path))
			}

			// 合并所有的配置文件到内存当中
			name := ns[1]
			val := v.GetStringMap(name)
			val1 := UnMarshal(v, path)
			if val != nil {
				xerror.Panic(mergo.Map(&val, val1, mergo.WithOverride, mergo.WithTypeCheck))
				val1 = val
			}
			v.Set(name, val1)

			return nil
		}))
	}

	return nil
}

func init() {
	// watch
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		GetCfg().WatchConfig()

		if golug_env.IsDev() || golug_env.IsTest() {
			golug_watcher.AddWatcher(file.NewWatcher(CfgType, CfgName, func(path string) map[string]interface{} {
				return UnMarshal(GetCfg().Viper, path)
			}))
		}

		if GetCfg().GetBool("watcher.configs.etcd.enabled") {
			name := GetCfg().GetString("watcher.configs.etcd.driver")
			cfg, ok := golug_etcd.GetCfg().Configs[name]
			if ok && cfg.Enabled {
				c := xerror.PanicErr(golug_etcd.GetClient(name)).(*clientv3.Client)
				golug_watcher.AddWatcher(etcd.NewWatcher(golug_env.Project, c))
			}
		}

		golug_watcher.Start()
	}))

	xerror.Exit(dix_run.WithBeforeStop(func(ctx *dix_run.BeforeStopCtx) { golug_watcher.Close() }))
}
