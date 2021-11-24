package config

import (
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/iox"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/x/typex"
	"github.com/pubgo/xerror"
	"github.com/spf13/viper"

	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/runenv"
)

var _ Config = (*configImpl)(nil)

type configImpl struct {
	rw   sync.RWMutex
	v    *viper.Viper
	init bool
}

func (t *configImpl) check() {
	if t.init {
		return
	}

	panic("please init config")
}

func (t *configImpl) All() map[string]interface{} {
	t.check()

	t.rw.RLock()
	defer t.rw.RUnlock()

	return t.v.AllSettings()
}

func (t *configImpl) MergeConfig(in io.Reader) error {
	t.check()

	t.rw.Lock()
	defer t.rw.Unlock()

	return t.v.MergeConfig(in)
}

func (t *configImpl) AllKeys() []string {
	t.check()

	t.rw.RLock()
	defer t.rw.RUnlock()

	return t.v.AllKeys()
}

func (t *configImpl) ConfigPath() string {
	t.check()

	t.rw.RLock()
	defer t.rw.RUnlock()

	return t.v.ConfigFileUsed()
}

func (t *configImpl) GetMap(key string) map[string]interface{} {
	t.check()

	t.rw.RLock()
	defer t.rw.RUnlock()

	return t.v.GetStringMap(key)
}

func (t *configImpl) Get(key string) interface{} {
	t.check()

	t.rw.RLock()
	defer t.rw.RUnlock()

	return t.v.Get(key)
}

func (t *configImpl) GetString(key string) string {
	t.check()

	t.rw.RLock()
	defer t.rw.RUnlock()

	return t.v.GetString(key)
}

func (t *configImpl) Set(key string, value interface{}) {
	t.check()

	t.rw.Lock()
	defer t.rw.Unlock()

	t.v.Set(key, value)
}

func (t *configImpl) UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	t.check()

	t.rw.RLock()
	defer t.rw.RUnlock()

	return t.v.UnmarshalKey(key, rawVal, opts...)
}

func (t *configImpl) Decode(name string, fn interface{}) (err error) {
	defer xerror.RespErr(&err)

	t.check()

	xerror.Assert(name == "" || fn == nil, "[name,fn] should not be nil")
	if t.Get(name) == nil {
		return ErrKeyNotFound
	}

	vfn := reflect.ValueOf(fn)
	switch vfn.Type().Kind() {
	case reflect.Func: // func(cfg *Cfg)
		xerror.Assert(vfn.Type().NumIn() != 1, "[fn] input num should be 1")

		mthIn := reflect.New(vfn.Type().In(0).Elem())
		ret := fx.WrapRaw(t.UnmarshalKey)(name, mthIn,
			func(cfg *mapstructure.DecoderConfig) { cfg.TagName = "json" })

		if !ret[0].IsNil() {
			xerror.PanicF(ret[0].Interface().(error),
				"config key [%s] decode error", name)
		}

		vfn.Call(typex.ValueOf(mthIn))
	case reflect.Ptr:
		return xerror.WrapF(t.UnmarshalKey(name, fn,
			func(cfg *mapstructure.DecoderConfig) { cfg.TagName = "json" },
		), "config key [%s] decode error", name)
	default:
		return xerror.Fmt("[fn] type error,name=>%s, refer=>%#v", name, fn)
	}

	return nil
}

// Init 处理所有的配置,环境变量和flag
// 配置顺序, 默认值->环境变量->配置文件->flag
// 配置文件中可以设置环境变量
// flag可以指定配置文件位置
// 始化配置文件
func (t *configImpl) Init() (err error) {
	defer func() { t.init = err == nil }()

	defer xerror.RespErr(&err)

	t.rw.Lock()
	defer t.rw.Unlock()

	xerror.Assert(!runenv.CheckMode(), "mode(%s) not match in all(%v)", runenv.Mode, runenv.RunmodeValue)

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
	xerror.PanicF(t.initWithDir(v), "config file load error")
	Home = filepath.Dir(filepath.Dir(v.ConfigFileUsed()))

	dt := xerror.PanicStr(iox.ReadText(v.ConfigFileUsed()))

	// 处理配置文件中的环境变量
	dt = env.Expand(dt)

	// 重新加载配置
	xerror.Panic(v.MergeConfig(strings.NewReader(dt)))

	// 加载自定义配置
	xerror.Panic(t.initApp(v))
	return nil
}

func (t *configImpl) addConfigPath(v *viper.Viper, in string) bool {
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

func (t *configImpl) initWithCfg(v *viper.Viper) bool {
	if CfgPath == "" {
		return false
	}

	xerror.Assert(pathutil.IsNotExist(CfgPath), "config file not found, path:%s", CfgPath)

	v.SetConfigFile(CfgPath)

	xerror.PanicF(v.ReadInConfig(), "config load error, path:%s", CfgPath)

	return true
}

func (t *configImpl) initWithDir(v *viper.Viper) (err error) {
	defer xerror.RespErr(&err)

	// 指定配置文件
	if t.initWithCfg(v) {
		return
	}

	// 检查配置是否存在
	if v.ReadInConfig() == nil {
		return nil
	}

	var pathList = strListMap(getPathList(), func(str string) string { return filepath.Join(str, ".lava", CfgName) })
	for i := range pathList {
		if t.addConfigPath(v, pathList[i]) {
			return
		}
	}

	return xerror.Wrap(v.ReadInConfig())
}

// 监控配置中的app自定义配置
func (t *configImpl) initApp(v *viper.Viper) error {
	// .lava/config/config.dev.yaml
	var path = filepath.Join(Home, "config", fmt.Sprintf("%s.%s.%s", CfgName, runenv.Mode, CfgType))
	if !pathutil.IsExist(path) {
		return nil
	}

	// 读取配置
	dt := xerror.PanicStr(iox.ReadText(path))

	// 处理环境变量
	dt = env.Expand(dt)

	c := make(map[string]interface{})
	xerror.Panic(unmarshalReader(v, strings.NewReader(dt), c))

	// 合并自定义配置
	xerror.Panic(v.MergeConfigMap(c))
	return nil
}

func getPathList() (paths []string) {
	var wd = xerror.PanicStr(filepath.Abs("./"))
	for {
		if wd == "/" {
			break
		}

		paths = append(paths, wd)
		wd = filepath.Dir(wd)
	}
	return
}

func strListMap(strList []string, fn func(str string) string) []string {
	for i := range strList {
		strList[i] = fn(strList[i])
	}
	return strList
}
