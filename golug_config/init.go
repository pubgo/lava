package golug_config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pubgo/dix"
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/pkg/golug_utils"
	"github.com/pubgo/xerror"
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

	v.AddConfigPath(fmt.Sprintf("/etc/%s/%s/", golug_app.Domain, golug_app.Project))
	v.AddConfigPath(fmt.Sprintf("$HOME/.%s/%s", golug_app.Domain, golug_app.Project))

	// 监控Home工作目录
	_home := xerror.PanicErr(homedir.Dir()).(string)
	v.AddConfigPath(filepath.Join(_home, "."+golug_app.Project, CfgName))
	return nil
}

func InitOtherConfig() (err error) {
	defer xerror.RespErr(&err)

	v := GetCfg()

	return xerror.Wrap(filepath.Walk(filepath.Dir(v.ConfigFileUsed()), func(path string, info os.FileInfo, err error) error {
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

		fmt.Println(info.Name(), CfgName+"."+CfgType)
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
}

func Init() (err error) {
	defer xerror.RespErr(&err)

	cfg = &Config{Viper: viper.New()}
	v := cfg.Viper

	// 配置文件名字和类型
	v.SetConfigType(CfgType)
	v.SetConfigName(CfgName)

	// config 路径
	v.AddConfigPath(".")
	v.SetEnvPrefix(golug_app.Domain)
	v.SetEnvKeyReplacer(strings.NewReplacer("_", ".", "-", ".", "/", "."))
	v.AutomaticEnv()

	// 监控默认配置
	v.AddConfigPath(filepath.Join(golug_app.Home, CfgName))

	// 监控当前工作目录
	_pwd := xerror.PanicStr(filepath.Abs(filepath.Dir("")))
	v.AddConfigPath(filepath.Join(_pwd, CfgName))

	// 检查配置文件是否存在
	if err := v.ReadInConfig(); err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			return nil
		}
		return xerror.WrapF(err, "read config failed")
	}

	// 获取配置文件所在目录
	CfgPath = xerror.PanicStr(filepath.Abs(v.ConfigFileUsed()))
	golug_app.Home = filepath.Dir(filepath.Dir(CfgPath))

	return nil
}

func IsExist() bool {
	return GetCfg().ReadInConfig() == nil
}

type Ctx struct{ dix.Model }

func Trigger() error     { return xerror.Wrap(dix.Dix(Ctx{})) }
func On(fn func(_ *Ctx)) { xerror.Next().Panic(dix.Dix(fn)) }
