package config

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/logx"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/xerr"
	"github.com/pubgo/x/iox"
	"github.com/pubgo/x/pathutil"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/internal/envs"
	"github.com/pubgo/lava/internal/pkg/env"
	"github.com/pubgo/lava/internal/pkg/merge"
	"github.com/pubgo/lava/internal/pkg/reflectx"
	"github.com/pubgo/lava/internal/pkg/typex"
)

const (
	FileType = "yaml"
	FileName = "config"
)

var (
	CfgDir  string
	CfgPath string
)

// Init 处理所有的配置,环境变量和flag
// 配置顺序, 默认值->环境变量->配置文件->flag
// 配置文件中可以设置环境变量
// flag可以指定配置文件位置
// 始化配置文件
func newCfg() *configImpl {
	defer recovery.Exit()

	var t = &configImpl{v: viper.New()}
	// 配置处理
	v := t.v

	// 配置文件名字和类型
	v.SetConfigType(FileType)
	v.SetConfigName(FileName)
	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_", "/", "_"))
	v.SetEnvPrefix(strings.ToUpper(envs.EnvPrefix))
	v.AutomaticEnv()

	// 初始化框架, 加载环境变量, 加载本地配置
	// 初始化完毕所有的配置以及外部配置以及相关的参数和变量
	// 然后获取配置了
	t.initCfg(v)
	CfgPath = v.ConfigFileUsed()
	CfgDir = filepath.Dir(v.ConfigFileUsed())
	assert.Must(env.Set(consts.EnvHome, CfgDir))
	assert.Must(t.LoadPath(CfgPath))

	// 加载自定义配置
	assert.Must(t.loadCustomCfg())
	return t
}

var _ Config = (*configImpl)(nil)

type configImpl struct {
	rw sync.RWMutex
	v  *viper.Viper
}

func (t *configImpl) loadCustomCfg() (err error) {
	defer recovery.Err(&err)

	var cfg App
	assert.Must(t.UnmarshalKey("app", &cfg))

	for _, path := range cfg.Resources {
		assert.Must(t.LoadPath(filepath.Join(CfgDir, path)))
	}
	return nil
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
	return t.v.UnmarshalKey(key, rawVal, append(opts, func(c *mapstructure.DecoderConfig) {
		if c.TagName == "" {
			c.TagName = FileType
		}
	})...)
}

// Decode decode config to map[string]*struct
func (t *configImpl) Decode(name string, cfgMap interface{}) (gErr error) {
	defer recovery.Err(&gErr, func(err xerr.XErr) xerr.XErr {
		return err.WrapF("name=%s, cfgMap=%#v", name, cfgMap)
	})

	assert.If(name == "" || cfgMap == nil, "[name,cfgMap] should not be nil")
	assert.If(reflectx.Indirect(reflect.ValueOf(cfgMap)).Kind() != reflect.Map, "[cfgMap](%#v) should be map", cfgMap)
	assert.If(t.Get(name) == nil, "config(%s) key not found", name)

	var cfg *typex.RwMap
	for _, data := range cast.ToSlice(t.Get(name)) {
		var dm = assert.Must1(cast.ToStringMapE(data))
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

	assert.MustF(merge.MapStruct(cfgMap, cfg.Map()), "config key [%s] decode error", name)
	return
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

	assert.MustF(err, "read config failed, path:%s", in)
	return false
}

func (t *configImpl) initWithConfig(v *viper.Viper) bool {
	if CfgDir == "" {
		return false
	}

	assert.Assert(pathutil.IsNotExist(CfgDir), "config not found, path:%s", CfgDir)
	v.AddConfigPath(CfgDir)
	assert.MustF(v.ReadInConfig(), "config load error, config:%s", CfgDir)

	return true
}

func (t *configImpl) initCfg(v *viper.Viper) {

	// 指定配置文件
	if t.initWithConfig(v) {
		return
	}

	// 检查配置是否存在
	if v.ReadInConfig() == nil {
		return
	}

	var pathList = strMap(getPathList(), func(str string) string { return filepath.Join(str, "configs") })
	assert.If(len(pathList) == 0, "config path not found")

	for i := range pathList {
		if t.addConfigPath(pathList[i]) {
			return
		}
	}

	assert.Must(v.ReadInConfig())
}

// LoadPath 加载指定path的配置
func (t *configImpl) LoadPath(path string) (err error) {
	defer recovery.Err(&err)

	if !pathutil.IsExist(path) {
		return nil
	}

	logx.V(1).Info("load config path", "path", path)

	tmpl := assert.Must1(template.New("").Funcs(template.FuncMap{
		"upper": strings.ToUpper,
		"trim":  strings.TrimSpace,
		"env":   env.Get,
		"v":     t.v.GetString,
		"default": func(a string, b string) string {
			if strings.TrimSpace(b) == "" {
				return a
			}
			return b
		},
	}).Parse(assert.Must1(iox.ReadText(path))))

	var buf bytes.Buffer
	assert.Must(tmpl.Execute(&buf, map[string]string{}))

	// 合并配置
	assert.Must(t.v.MergeConfig(&buf))
	return
}
