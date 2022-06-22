package config

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/funk"
	"github.com/pubgo/x/iox"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/valyala/fasttemplate"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/internal/pkg/env"
	"github.com/pubgo/lava/internal/pkg/merge"
	"github.com/pubgo/lava/internal/pkg/reflectx"
	"github.com/pubgo/lava/internal/pkg/typex"
	"github.com/pubgo/lava/internal/pkg/utils"
)

var (
	CfgType   = "yaml"
	CfgName   = "config"
	CfgDir    = env.Get("app_cfg_home", "project_cfg_home", consts.EnvCfgHome)
	CfgPath   = filepath.Join("configs", "config", "config.yaml")
	EnvPrefix = utils.FirstNotEmpty(env.Get("cfg_env_prefix", "app_env_prefix", "project_env_prefix", consts.EnvCfgPrefix), "lava")
)

// Init 处理所有的配置,环境变量和flag
// 配置顺序, 默认值->环境变量->配置文件->flag
// 配置文件中可以设置环境变量
// flag可以指定配置文件位置
// 始化配置文件
func newCfg() *configImpl {
	var t = &configImpl{v: viper.New()}
	// 配置处理
	v := t.v

	// 配置文件名字和类型
	v.SetConfigType(CfgType)
	v.SetConfigName(CfgName)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.SetEnvPrefix(strings.ToUpper(EnvPrefix))
	v.AutomaticEnv()

	// 初始化框架, 加载环境变量, 加载本地配置
	// 初始化完毕所有的配置以及外部配置以及相关的参数和变量
	// 然后获取配置了
	xerror.PanicF(t.initCfg(v), "config file load error")
	CfgPath = v.ConfigFileUsed()
	CfgDir = filepath.Dir(filepath.Dir(v.ConfigFileUsed()))
	xerror.Panic(env.Set(consts.EnvCfgHome, CfgDir))
	t.LoadPath(CfgPath)

	// 加载自定义配置
	t.loadCustomCfg()
	return t
}

var _ Config = (*configImpl)(nil)

type configImpl struct {
	rw sync.RWMutex
	v  *viper.Viper
}

func (t *configImpl) loadCustomCfg() {
	var path = filepath.Dir(t.v.ConfigFileUsed())
	funk.MustMsg(filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), "."+CfgType) {
			return nil
		}

		if info.Name() == CfgName+"."+CfgType {
			return nil
		}

		t.LoadPath(path)
		return nil
	}), "walk path failed, path=%s", path)
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
	defer xerror.RecoverErr(&err, func(err xerror.XErr) xerror.XErr {
		return err.WrapF("name=%s, cfgMap=%#v", name, cfgMap)
	})

	xerror.Assert(name == "" || cfgMap == nil, "[name,cfgMap] should not be nil")
	xerror.Assert(reflectx.Indirect(reflect.ValueOf(cfgMap)).Kind() != reflect.Map, "[cfgMap](%#v) should be map", cfgMap)
	xerror.Assert(t.Get(name) == nil, "config(%s) key not found", name)

	var cfg *typex.RwMap
	for _, data := range cast.ToSlice(t.Get(name)) {
		var dm = xerror.PanicErr(cast.ToStringMapE(data)).(map[string]interface{})
		resId := getPkgId(dm)

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

func (t *configImpl) addConfigPath(in string) bool {
	t.v.AddConfigPath(in)
	err := t.v.ReadInConfig()
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

func (t *configImpl) initWithDir(v *viper.Viper) bool {
	if CfgDir == "" {
		return false
	}

	funk.Assert(pathutil.IsNotExist(CfgDir), "config dir not found, path:%s", CfgPath)
	v.AddConfigPath(filepath.Join(CfgDir, CfgName))
	funk.MustMsg(v.ReadInConfig(), "config load error, dir:%s", CfgDir)

	return true
}

func (t *configImpl) initCfg(v *viper.Viper) (err error) {
	defer xerror.RecoverErr(&err)

	// 指定配置目录
	if t.initWithDir(v) {
		return
	}

	// 检查配置是否存在
	if v.ReadInConfig() == nil {
		return nil
	}

	var pathList = strMap(getPathList(), func(str string) string { return filepath.Join(str, "configs", CfgName) })
	xerror.Assert(len(pathList) == 0, "config path not found")

	for i := range pathList {
		if t.addConfigPath(pathList[i]) {
			return
		}
	}

	return xerror.Wrap(v.ReadInConfig())
}

// LoadPath 加载指定path的配置
func (t *configImpl) LoadPath(path string) {
	if !pathutil.IsExist(path) {
		return
	}

	fmt.Printf("load config %s\n", path)

	tmp, err := fasttemplate.NewTemplate(xerror.PanicStr(iox.ReadText(path)), "{{", "}}")
	xerror.Panic(err, "unexpected error when parsing template")

	// 重新加载配置
	xerror.Panic(t.v.MergeConfig(strings.NewReader(tmp.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		tag = strings.TrimSpace(tag)
		// 处理配置中的环境变量
		if strings.HasPrefix(tag, "$") {
			return w.Write([]byte(env.Get(tag)))
		}

		return w.Write([]byte(t.v.GetString(tag)))
	}))))
}
