package config

import (
	"fmt"
	"github.com/pubgo/funk/result"
	"os"
	"path/filepath"
	"reflect"

	"github.com/imdario/mergo"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
)

func getComponentName(m map[string]interface{}) string {
	if m == nil || len(m) == 0 {
		return defaultComponentKey
	}

	val, ok := m[componentConfigKey]
	if !ok || val == nil {
		return defaultComponentKey
	}

	return fmt.Sprintf("%v", val)
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

func getCfgData() interface{} {
	cfg := New()
	return map[string]any{
		"cfg_type":   defaultConfigType,
		"cfg_name":   defaultConfigName,
		"home":       CfgDir,
		"cfg_path":   CfgPath,
		"all_key":    cfg.AllKeys(),
		"all_config": cfg.All(),
	}
}

func Load[T any]() T {
	c := New()
	var cfg T
	assert.Must(c.Unmarshal(&cfg))
	return cfg
}

func Merge[A any, B any](dst A, src ...B) (ret result.Result[*A]) {
	for i := range src {
		err := mergo.Merge(&dst, src[i], mergo.WithOverride, mergo.WithAppendSlice, mergo.WithTransformers(new(transformer)))
		if err != nil {
			return ret.WithErr(errors.WrapTag(err,
				errors.T("dst", dst),
				errors.T("src", src[i]),
			))
		}
	}
	return ret.WithVal(&dst)
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
					}
				})
			}
		}

		return nil
	}
}
