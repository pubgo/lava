package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/iox"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/runtime"
)

var errType = reflect.TypeOf((*error)(nil)).Elem()

var _ config_type.Config = (*configImpl)(nil)

type configImpl struct {
	rw   sync.RWMutex
	v    *viper.Viper
	init bool
}

func (t *configImpl) All() map[string]interface{} {
	t.rw.RLock()
	defer t.rw.RUnlock()

	return t.v.AllSettings()
}

func (t *configImpl) MergeConfig(in io.Reader) error {
	t.rw.Lock()
	defer t.rw.Unlock()

	return t.v.MergeConfig(in)
}

func (t *configImpl) AllKeys() []string {
	t.rw.RLock()
	defer t.rw.RUnlock()

	return t.v.AllKeys()
}

func (t *configImpl) GetMap(keys ...string) config_type.CfgMap {
	t.rw.RLock()
	defer t.rw.RUnlock()

	key := strings.Trim(strings.Join(keys, "."), ".")
	var val = t.v.Get(key)
	if val == nil {
		return config_type.CfgMap{}
	}

	for _, data := range cast.ToSlice(val) {
		return cast.ToStringMap(data)
	}

	return t.v.GetStringMap(key)
}

func (t *configImpl) Get(key string) interface{} {
	t.rw.RLock()
	defer t.rw.RUnlock()

	return t.v.Get(key)
}

func (t *configImpl) GetString(key string) string {
	t.rw.RLock()
	defer t.rw.RUnlock()

	return t.v.GetString(key)
}

func (t *configImpl) Set(key string, value interface{}) {
	t.rw.Lock()
	defer t.rw.Unlock()

	t.v.Set(key, value)
}

func (t *configImpl) Decode(name string, val interface{}) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(name == "" || val == nil, "[name,val] should not be nil")
	if t.Get(name) == nil {
		return ErrKeyNotFound
	}

	vfn := reflect.ValueOf(val)
	switch vfn.Type().Kind() {
	case reflect.Func: // func(cfg *Struct)error
		xerror.Assert(vfn.Type().NumIn() != 1, "[val] input num should be 1")
		xerror.Assert(vfn.Type().NumOut() != 1, "[val] output num should be 1")
		xerror.Assert(!vfn.Type().Out(0).Implements(errType), "[val] output should be error type")

		mthIn := reflect.New(vfn.Type().In(0).Elem())
		ret := fx.WrapRaw(t.v.UnmarshalKey)(name, mthIn)

		if !ret[0].IsNil() {
			xerror.PanicF(ret[0].Interface().(error), "config key [%s] decode error", name)
		}

		vfn.Call([]reflect.Value{mthIn})
	case reflect.Ptr:
		return xerror.WrapF(t.v.UnmarshalKey(name, val), "config key [%s] decode error", name)
	default:
		return xerror.Fmt("[val] type error,name=>%s, refer=>%#v", name, val)
	}

	return nil
}

// Init 处理所有的配置,环境变量和flag
// 配置顺序, 默认值->环境变量->配置文件->flag
// 配置文件中可以设置环境变量
// flag可以指定配置文件位置
// 始化配置文件
func (t *configImpl) initCfg() {
	defer func() { t.init = true }()
	defer xerror.RespExit()

	t.rw.Lock()
	defer t.rw.Unlock()

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
	CfgPath = v.ConfigFileUsed()
	os.Setenv("data", Home)

	dt := xerror.PanicStr(iox.ReadText(v.ConfigFileUsed()))

	// 处理配置中的环境变量
	dt = env.Expand(dt)

	// 重新加载配置
	xerror.Panic(v.MergeConfig(strings.NewReader(dt)))

	// 加载自定义配置
	xerror.Panic(t.initApp(v))
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

	var pathList = strMap(getPathList(), func(str string) string { return filepath.Join(str, ".lava", CfgName) })
	xerror.Assert(len(pathList) == 0, "paths is ")

	for i := range pathList {
		if t.addConfigPath(v, pathList[i]) {
			return
		}
	}

	return xerror.Wrap(v.ReadInConfig())
}

// 监控配置中的app自定义配置
func (t *configImpl) initApp(v *viper.Viper) error {
	// .lava/config/config.[env].yaml
	var path = filepath.Join(filepath.Dir(CfgPath), fmt.Sprintf("%s.%s.%s", CfgName, runtime.Mode, CfgType))
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
