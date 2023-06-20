package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/imdario/mergo"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/pathutil"
	"gopkg.in/yaml.v3"

	"github.com/pubgo/lava/core/vars"
)

func getConfigPath(name, typ string, configDir ...string) (config string, dir string) {
	if len(configDir) == 0 {
		configDir = append(configDir, "./", defaultConfigPath)
	}

	if name == "" {
		name = defaultConfigName
	}

	if typ == "" {
		typ = defaultConfigType
	}

	var configName = fmt.Sprintf("%s.%s", name, typ)
	var notFoundPath []string
	for _, path := range getPathList() {
		for _, dir := range configDir {
			var configPath = filepath.Join(path, dir, configName)
			if pathutil.IsNotExist(configPath) {
				notFoundPath = append(notFoundPath, configPath)
			} else {
				return configPath, filepath.Dir(configPath)
			}
		}
	}

	log.Panicf("config not found in: %v\n", notFoundPath)

	return "", ""
}

// getPathList 递归得到当前目录到跟目录中所有的目录路径
//
//	paths: [./, ../, ../../, ..., /]
func getPathList() (paths []string) {
	wd := assert.Must1(filepath.Abs(""))
	for len(wd) > 0 && !os.IsPathSeparator(wd[len(wd)-1]) {
		paths = append(paths, wd)
		wd = filepath.Dir(wd)
	}
	return
}

func strMap(strList []string, fn func(str string) string) []string {
	for i := range strList {
		strList[i] = fn(strList[i])
	}
	return strList
}

func Load[T any]() T {
	configPath, configDir := getConfigPath("", "")
	configBytes := assert.Must1(os.ReadFile(configPath))

	var cfg T
	assert.Must(yaml.Unmarshal(configBytes, &cfg))

	var res Resources
	assert.Must(yaml.Unmarshal(configBytes, &res))

	var cfgList []T
	for _, resPath := range res.Resources {
		var cfg1 T
		resAbsPath := filepath.Join(configDir, resPath)
		if pathutil.IsNotExist(resAbsPath) {
			log.Panicln("resources config path not found:", resAbsPath)
		}
		resBytes := assert.Must1(os.ReadFile(resAbsPath))
		assert.Must(yaml.Unmarshal(resBytes, &cfg1))
		cfgList = append(cfgList, cfg1)
	}

	assert.Must(Merge(&cfg, cfgList...))

	vars.RegisterValue("config", map[string]any{
		"config_type": defaultConfigType,
		"config_name": defaultConfigName,
		"config_path": CfgPath,
		"config_dir":  CfgDir,
		"config_data": cfg,
	})

	return cfg
}

func Merge[A any, B any](dst *A, src ...B) error {
	for i := range src {
		err := mergo.Merge(dst, src[i], mergo.WithOverride, mergo.WithAppendSlice, mergo.WithTransformers(new(transformer)))
		if err != nil {
			return errors.WrapTag(err,
				errors.T("dst_type", reflect.TypeOf(dst).String()),
				errors.T("dst", dst),
				errors.T("src_type", reflect.TypeOf(src[i]).String()),
				errors.T("src", src[i]),
			)
		}
	}
	return nil
}

type transformer struct{}

func (s *transformer) Transformer(t reflect.Type) func(dst, src reflect.Value) error {
	if t == nil || t.Kind() != reflect.Slice {
		return nil
	}

	if !t.Elem().Implements(reflect.TypeOf((*NamedConfig)(nil)).Elem()) {
		return nil
	}

	return func(dst, src reflect.Value) error {
		if !src.IsValid() || src.IsNil() {
			return nil
		}

		var dstMap = make(map[string]NamedConfig)
		for i := 0; i < dst.Len(); i++ {
			c := dst.Index(i).Interface().(NamedConfig)
			dstMap[c.ConfigUniqueName()] = c
		}

		for i := 0; i < src.Len(); i++ {
			c := src.Index(i).Interface().(NamedConfig)
			if dstMap[c.ConfigUniqueName()] == nil {
				dst = reflect.Append(dst, reflect.ValueOf(c))
				dstMap[c.ConfigUniqueName()] = c
				continue
			}

			d := dstMap[c.ConfigUniqueName()]
			err := mergo.Merge(d, c, mergo.WithOverride, mergo.WithAppendSlice, mergo.WithTransformers(new(transformer)))
			if err != nil {
				return errors.WrapFn(err, func() errors.Tags {
					return errors.Tags{
						errors.T("dst", d),
						errors.T("src", c),
						errors.T("dst-type", reflect.TypeOf(d).String()),
						errors.T("src-type", reflect.TypeOf(c).String()),
					}
				})
			}
		}

		return nil
	}
}
