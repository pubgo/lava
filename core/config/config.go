package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/a8m/envsubst"
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/merge"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/typex"
	"github.com/pubgo/funk/version"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// New 处理所有的配置,环境变量和flag
// 配置顺序, 默认值->环境变量->配置文件->flag
// 配置文件中可以设置环境变量
// flag可以指定配置文件位置
// 始化配置文件
func New() Config {
	defer recovery.Exit()

	// 配置处理
	v := viper.New()
	v.SetEnvKeyReplacer(replacer)

	envPrefix := strings.ToUpper(replacer.Replace(version.Project()))
	log.Info().Str("env_prefix", envPrefix).Msg("set config env prefix")
	v.SetEnvPrefix(envPrefix)

	v.SetConfigName(defaultConfigName)
	v.SetConfigType(defaultConfigType)
	v.AutomaticEnv()
	v.AddConfigPath(".")
	v.AddConfigPath(defaultConfigPath)

	t := &configImpl{v: v}
	// 初始化框架, 加载环境变量, 加载本地配置
	// 初始化完毕所有的配置以及外部配置以及相关的参数和变量
	// 然后获取配置了
	t.initCfg()
	CfgPath = v.ConfigFileUsed()
	CfgDir = filepath.Dir(v.ConfigFileUsed())
	t.loadPath(CfgPath)

	// 加载自定义配置
	t.loadCustomCfg()
	log.Info().Any("metadata", map[string]any{
		"cfg_type": defaultConfigType,
		"cfg_name": defaultConfigName,
		"home":     CfgDir,
		"cfg_path": CfgPath,
	}).Msg("config metadata")
	log.Debug().Any("data", t.All()).Msg("config settings")
	return t
}

type configImpl struct {
	v *viper.Viper
}

func (t *configImpl) loadCustomCfg() {
	includes := t.v.GetStringSlice(includeConfigName)
	for _, path := range includes {
		t.loadPath(filepath.Join(CfgDir, path))
	}
}

func (t *configImpl) All() map[string]interface{} {
	return t.v.AllSettings()
}

func (t *configImpl) AllKeys() []string {
	return t.v.AllKeys()
}

func (t *configImpl) Get(key string) interface{} {
	return t.v.Get(key)
}

func (t *configImpl) GetString(key string) string {
	return t.v.GetString(key)
}

func (t *configImpl) Set(key string, value interface{}) {
	t.v.Set(key, value)
}

func (t *configImpl) UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return t.v.UnmarshalKey(key, rawVal, append(opts, func(c *mapstructure.DecoderConfig) {
		if c.TagName == "" {
			c.TagName = defaultConfigType
		}
	})...)
}

func (t *configImpl) Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return t.v.Unmarshal(rawVal, append(opts, func(c *mapstructure.DecoderConfig) {
		if c.TagName == "" {
			c.TagName = defaultConfigType
		}
	})...)
}

// DecodeComponent decode component config to map[string]*struct
func (t *configImpl) DecodeComponent(name string, cfgMap interface{}) (gErr error) {
	defer recovery.Err(&gErr, func(err error) error {
		return errors.WrapTags(err, errors.Tags{
			"name":   name,
			"cfgMap": cfgMap,
		})
	})

	assert.If(name == "" || cfgMap == nil, "name,cfgMap params should not be nil")
	assert.If(reflect.Indirect(reflect.ValueOf(cfgMap)).Kind() != reflect.Map, "cfgMap param should be map type")
	assert.If(t.Get(name) == nil, "config value not found")

	var cfg *typex.RwMap
	for _, data := range cast.ToSlice(t.Get(name)) {
		if cfg == nil {
			cfg = &typex.RwMap{}
		}

		dm := assert.Must1(cast.ToStringMapE(data))
		componentName := getComponentName(dm)
		assert.Err(cfg.Has(componentName), errors.Err{
			Msg:    "component name already exists",
			Detail: fmt.Sprintf("key=%s component_name=%s", name, componentName),
		})

		cfg.Set(componentName, dm)
	}

	if cfg == nil {
		cfg = &typex.RwMap{}
		cfg.Set(defaultComponentKey, t.Get(name))
	}

	return merge.MapStruct(cfgMap, cfg.Map()).Err()
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

	assert.MustF(err, "failed to read config, path=%s", in)
	return false
}

func (t *configImpl) initWithConfig() bool {
	if CfgDir == "" {
		return false
	}

	assert.If(pathutil.IsDir(CfgDir) && pathutil.IsNotExist(CfgDir), "config not found, path=%s", CfgDir)
	t.v.AddConfigPath(CfgDir)
	assert.MustF(t.v.ReadInConfig(), "failed to load config, path=%s", CfgDir)
	return true
}

func (t *configImpl) initCfg() {
	// 指定配置目录
	if t.initWithConfig() {
		return
	}

	// 检查配置是否存在
	if t.v.ReadInConfig() == nil {
		return
	}

	pathList := strMap(getPathList(), func(str string) string { return filepath.Join(str, "configs") })
	assert.If(len(pathList) == 0, "config path not found")

	for i := range pathList {
		if t.addConfigPath(pathList[i]) {
			return
		}
	}

	assert.MustF(t.v.ReadInConfig(), "path=%s", t.v.ConfigFileUsed())
}

// loadPath 加载指定path的配置
func (t *configImpl) loadPath(path string) {
	defer recovery.Exit()

	fileStat := assert.Must1(os.Stat(path))
	if fileStat.IsDir() {
		assert.Must(filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if !strings.HasSuffix(info.Name(), "."+defaultConfigType) {
				return nil
			}

			t.loadPath(path)
			return nil
		}))
		return
	}

	assert.If(pathutil.IsNotExist(path), "path not found, path=%s", path)
	log.Info().Str("path", path).Msgf("load config path")

	cfgData := string(assert.Must1(os.ReadFile(path)))
	cfgData = assert.Must1(envsubst.String(cfgData))
	assert.Must(t.v.MergeConfig(strings.NewReader(cfgData)))
	return
}
