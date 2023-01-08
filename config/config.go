package config

import (
	"fmt"
	"github.com/pubgo/funk/logx"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/a8m/envsubst"
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/xerr"
	"github.com/pubgo/x/iox"
	"github.com/pubgo/x/pathutil"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/pkg/reflectx"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/version"
)

const (
	FileType = "yaml"
	FileName = "config"
)

var (
	CfgDir  string
	CfgPath string
)

const (
	defaultConfigName = "config"
	defaultConfigType = "yaml"
	defaultConfigPath = "./configs"
)

// New 处理所有的配置,环境变量和flag
// 配置顺序, 默认值->环境变量->配置文件->flag
// 配置文件中可以设置环境变量
// flag可以指定配置文件位置
// 始化配置文件
func New() Config {
	defer recovery.Exit()

	replacer := strings.NewReplacer(".", "_", "-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix(version.Project())
	viper.AutomaticEnv()

	viper.SetConfigName(defaultConfigName)
	viper.SetConfigType(defaultConfigType)
	viper.AddConfigPath(defaultConfigPath)

	var t = &configImpl{v: viper.New()}
	// 配置处理
	v := t.v

	// 配置文件名字和类型
	v.SetConfigType(FileType)
	v.SetConfigName(FileName)

	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_", "/", "_"))
	v.SetEnvPrefix(strings.ToUpper(version.Project()))
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

type configImpl struct {
	v *viper.Viper
}

func (t *configImpl) loadCustomCfg() (err error) {
	defer recovery.Err(&err)

	var includes = t.v.GetStringSlice("includes")
	for _, path := range includes {
		assert.Must(t.LoadPath(filepath.Join(CfgDir, path)))
	}
	return nil
}

func (t *configImpl) All() map[string]interface{} {
	return t.v.AllSettings()
}

func (t *configImpl) MergeConfig(in io.Reader) error {
	return t.v.MergeConfig(in)
}

func (t *configImpl) AllKeys() []string {
	return t.v.AllKeys()
}

func (t *configImpl) GetMap(keys ...string) CfgMap {
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
			c.TagName = FileType
		}
	})...)
}

func (t *configImpl) Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	return t.v.Unmarshal(rawVal, append(opts, func(c *mapstructure.DecoderConfig) {
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

	merge.MapStruct(cfgMap, cfg.Map()).Unwrap(func(err result.Error) result.Error {
		return err.WrapF("config key [%s] decode error", name)
	})
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

	// 指定配置目录
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

	fi := assert.Must1(os.Stat(path))
	if fi.IsDir() {
		assert.Must(filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if !strings.HasSuffix(info.Name(), "."+FileType) {
				return nil
			}

			return t.LoadPath(path)
		}))
		return
	}

	logx.V(1).Info("load config path", "path", path)

	var subCfgData = assert.Must1(iox.ReadText(path))
	subCfgData = assert.Must1(envsubst.String(subCfgData))
	assert.Must(t.v.MergeConfig(strings.NewReader(subCfgData)))
	return
}
