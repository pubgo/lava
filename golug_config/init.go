package golug_config

import (
	"os"
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

	CfgPath = xerror.PanicStr(filepath.Abs(CfgPath))
	CfgPath = xerror.PanicStr(filepath.EvalSymlinks(CfgPath))
	CfgType = filepath.Ext(CfgPath)
	CfgName = strings.TrimSuffix(filepath.Base(CfgPath), "."+CfgType)
	GetCfg().SetConfigFile(CfgPath)
	golug_app.Home = filepath.Dir(filepath.Dir(CfgPath))
	return nil
}

func InitProject() (err error) {
	defer xerror.RespErr(&err)

	v := GetCfg()

	// etc目录
	v.AddConfigPath(filepath.Join("/etc", golug_app.Domain, golug_app.Project, CfgName))

	// 监控Home工作目录
	home := xerror.PanicErr(homedir.Dir()).(string)
	v.AddConfigPath(filepath.Join(home, "."+golug_app.Project, CfgName))
	v.AddConfigPath(filepath.Join(home, "."+golug_app.Domain, golug_app.Project, CfgName))
	return nil
}

// 监控配置中的其他配置文件
func InitOtherConfig() (err error) {
	defer xerror.RespErr(&err)

	v := GetCfg()

	// 处理独立的组件的配置, config.nsq.yaml, config.mysql.yaml
	xerror.Panic(filepath.Walk(filepath.Dir(v.ConfigFileUsed()), func(path string, info os.FileInfo, err error) error {
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
			return nil
		}

		// 合并所有的配置文件到内存当中
		name := ns[1]
		// 获取config中默认的配置
		val := v.GetStringMap(name)
		// 从自定义文件中解析配置
		val1 := UnMarshal(path)
		if val != nil {
			// 合并配置
			golug_utils.Mergo(&val, val1)
			val1 = val
		}
		v.Set(name, val1)

		return nil
	}))

	// 处理独立的app配置, app配置是使用方的自定义配置, config里面的是默认配置
	var appFile = filepath.Join(filepath.Dir(v.ConfigFileUsed()), "app."+CfgType)
	if golug_utils.PathExist(appFile) {
		for key, val := range UnMarshal(appFile) {
			v.Set(key, val)
		}
	}
	return nil
}

func Init() (err error) {
	defer xerror.RespErr(&err)

	v := viper.New()
	// 把环境变量的值设置到全局配置当中
	for key, val := range golug_env.List() {
		v.Set(key, val)
	}

	// 配置处理
	cfg = &Config{Viper: v}

	// 配置文件名字和类型
	v.SetConfigType(CfgType)
	v.SetConfigName(CfgName)

	// env 处理
	v.SetEnvPrefix(golug_app.Domain)
	v.SetEnvKeyReplacer(strings.NewReplacer("_", ".", "-", ".", "/", "."))
	v.AutomaticEnv()

	// config 路径
	// 当前目录
	v.AddConfigPath(filepath.Join("home", CfgName))

	// 检查配置文件是否存在
	if err := v.ReadInConfig(); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil
		}
		return xerror.WrapF(err, "read config failed")
	}

	// 获取配置文件所在目录
	CfgPath = xerror.PanicStr(filepath.Abs(v.ConfigFileUsed()))
	golug_app.Home = filepath.Dir(filepath.Dir(CfgPath))

	return nil
}
