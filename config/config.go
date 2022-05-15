package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/x/iox"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/valyala/fasttemplate"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/pkg/reflectx"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/runtime"
)

var (
	CfgType = "yaml"
	CfgName = "config"
	CfgDir  = strings.TrimSpace(env.Get("cfg_dir", "app_cfg_dir"))
	CfgPath = filepath.Join("configs", "config", "config.yaml")
)

// Init 处理所有的配置,环境变量和flag
// 配置顺序, 默认值->环境变量->配置文件->flag
// 配置文件中可以设置环境变量
// flag可以指定配置文件位置
// 始化配置文件
func newCfg() *configImpl {
	defer xerror.RespExit()

	var t = &configImpl{v: viper.New()}
	// 配置处理
	v := t.v

	// 配置文件名字和类型
	v.SetConfigType(CfgType)
	v.SetConfigName(CfgName)

	// 初始化框架, 加载环境变量, 加载本地配置
	// 初始化完毕所有的配置以及外部配置以及相关的参数和变量
	// 然后获取配置了
	xerror.PanicF(t.initWithDir(v), "config file load error")

	CfgPath = v.ConfigFileUsed()
	CfgDir = filepath.Dir(filepath.Dir(v.ConfigFileUsed()))
	xerror.Panic(os.Setenv(consts.EnvCfgHome, CfgDir))

	t.reload(t.v.ConfigFileUsed())

	// 加载自定义配置
	xerror.Panic(t.initApp(v))
	return t
}

var _ Config = (*configImpl)(nil)

type configImpl struct {
	rw sync.RWMutex
	v  *viper.Viper
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

func (t *configImpl) GetMap(keys ...string) CfgMap {
	t.rw.RLock()
	defer t.rw.RUnlock()

	key := strings.Trim(strings.Join(keys, "."), ".")
	var val = t.v.Get(key)
	if val == nil {
		return CfgMap{}
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

func (t *configImpl) UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	opts = append(opts, func(c *mapstructure.DecoderConfig) {
		if c.TagName == "" {
			c.TagName = CfgType
		}
	})
	return t.v.UnmarshalKey(key, rawVal, opts...)
}

// Decode decode config to map[string]*struct
func (t *configImpl) Decode(name string, cfgMap interface{}) (err error) {
	defer xerror.RespErr(&err)
	xerror.Assert(name == "" || cfgMap == nil, "[name,cfgMap] should not be nil")
	xerror.Assert(reflectx.Indirect(reflect.ValueOf(cfgMap)).Kind() != reflect.Map, "[cfgMap](%#v) should be map", cfgMap)
	xerror.Assert(t.Get(name) == nil, "config(%s) key not found", name)

	var cfg *typex.RwMap
	for _, data := range cast.ToSlice(t.Get(name)) {
		var dm = xerror.PanicErr(cast.ToStringMapE(data)).(map[string]interface{})
		resId := getResId(dm)

		if cfg == nil {
			cfg = &typex.RwMap{}
		}
		if _, ok := cfg.Load(resId); ok {
			panic(fmt.Errorf("res=>%s key=>%s,res key already exists", name, resId))
		}

		cfg.Set(resId, dm)
	}

	if cfg == nil {
		cfg = &typex.RwMap{}
		cfg.Set(consts.KeyDefault, t.Get(name))
	}

	return xerror.WrapF(merge.MapStruct(cfgMap, cfg.Map()), "config key [%s] decode error", name)
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
	if CfgDir == "" {
		return false
	}

	xerror.Assert(pathutil.IsNotExist(CfgDir), "config file not found, path:%s", CfgPath)
	v.AddConfigPath(filepath.Join(CfgDir, CfgName))
	xerror.PanicF(v.ReadInConfig(), "config load error, dir:%s", CfgDir)

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

	var pathList = strMap(getPathList(), func(str string) string { return filepath.Join(str, "configs", CfgName) })
	xerror.Assert(len(pathList) == 0, "pathList is zero")

	for i := range pathList {
		if t.addConfigPath(v, pathList[i]) {
			return
		}
	}

	return xerror.Wrap(v.ReadInConfig())
}

func (t *configImpl) reload(path string) {
	dt := xerror.PanicStr(iox.ReadText(path))

	// 处理配置中的环境变量
	dt = env.Expand(dt)

	// 重新加载配置
	xerror.Panic(t.v.MergeConfig(strings.NewReader(dt)))
	loadEnv(runtime.Project, t.v)
}

// 监控配置中的app自定义配置
func (t *configImpl) initApp(v *viper.Viper) error {
	// .lava/config/[env].yaml
	var path = filepath.Join(
		filepath.Dir(CfgPath),
		fmt.Sprintf("%s.%s.%s", CfgName, runtime.Mode, CfgType),
	)

	if !pathutil.IsExist(path) {
		return nil
	}

	t.reload(path)

	tmp, err := fasttemplate.NewTemplate(xerror.PanicStr(iox.ReadText(path)), "{{", "}}")
	xerror.Panic(err, "unexpected error when parsing template")
	xerror.Panic(v.MergeConfig(strings.NewReader(tmp.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		return w.Write([]byte(v.GetString(tag)))
	}))))
	return nil
}
