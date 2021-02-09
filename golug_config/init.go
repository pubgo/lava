package golug_config

import (
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pubgo/dix"
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/pkg/golug_utils"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// 默认的全局配置
var (
	Name    = "config"
	CfgType = "yaml"
	CfgName = "config"
	CfgPath = ""
	cfg     *Config
)

var trim = strings.TrimSpace

func IsExist() bool           { return GetCfg().ReadInConfig() == nil }
func Fire() error             { return xerror.Wrap(dix.Dix(GetCfg())) }
func On(fn func(cfg *Config)) { xerror.Panic(dix.Dix(fn)) }

func DefaultFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("cfg", pflag.PanicOnError)
	flags.StringVarP(&CfgPath, "cfg", "c", CfgPath, "config path")
	return flags
}

// 指定配置文件路径
func InitWithCfgPath() (err error) {
	defer xerror.RespErr(&err)

	v := GetCfg()

	CfgPath = xerror.PanicStr(filepath.Abs(CfgPath))
	CfgPath = xerror.PanicStr(filepath.EvalSymlinks(CfgPath))
	CfgType = filepath.Ext(CfgPath)
	CfgName = strings.TrimSuffix(filepath.Base(CfgPath), "."+CfgType)

	// 配置文件名字和类型
	v.SetConfigType(CfgType)
	v.SetConfigName(CfgName)
	v.SetConfigFile(CfgPath)

	return nil
}

func InitWithDir() (err error) {
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
	v.AddConfigPath(filepath.Join("/etc", golug_app.Domain, golug_app.Project, CfgName))

	// 监控Home工作目录
	home := xerror.PanicErr(homedir.Dir()).(string)
	v.AddConfigPath(filepath.Join(home, ".config", golug_app.Project, CfgName))
	v.AddConfigPath(filepath.Join(home, "."+golug_app.Domain, golug_app.Project, CfgName))

	if err := v.ReadInConfig(); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "not found") {
			return xerror.WrapF(err, "read config failed")
		}
	}
	return nil
}

// 监控配置中的其他配置文件
func InitApp() (err error) {
	defer xerror.RespErr(&err)

	v := GetCfg()

	// 处理独立的组件的配置, config.nsq.yaml, config.mysql.yaml
	appCfg := filepath.Join(filepath.Dir(v.ConfigFileUsed()), "app."+CfgType)
	if !golug_utils.PathExist(appCfg) {
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
			golug_utils.Mergo(&val, val2)
			val2 = val
		}
		v.Set(key, val2)
	}
	return nil
}

func Init() (err error) {
	defer xerror.RespErr(&err)

	// 配置处理
	cfg = &Config{Viper: viper.New()}

	v := cfg.Viper

	// env 处理
	v.SetEnvPrefix(golug_app.Domain)
	v.SetEnvKeyReplacer(strings.NewReplacer("_", ".", "-", ".", "/", "."))
	v.AutomaticEnv()

	// 把环境变量的值设置到全局配置当中
	for key, val := range golug_env.List() {
		v.Set(key, val)
	}

	// 配置文件名字和类型
	v.SetConfigType(CfgType)
	v.SetConfigName(CfgName)
	return nil
}

func InitHome() {
	// 获取配置文件所在目录
	CfgPath = xerror.PanicStr(filepath.Abs(GetCfg().ConfigFileUsed()))
	golug_app.Home = filepath.Dir(filepath.Dir(CfgPath))
}
