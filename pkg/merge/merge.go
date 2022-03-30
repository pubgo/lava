package merge

import (
	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/xerror"
)

type Option func(opts *copier.Option)

// Copy
// struct<->struct
// 各种类型结构体之间的field copy
func Copy(dst interface{}, src interface{}, opts ...Option) error {
	var optList copier.Option
	for i := range opts {
		opts[i](&optList)
	}

	return xerror.WrapF(copier.CopyWithOption(dst, src, optList), "\ndst: %#v\n\nsrc: %#v", dst, src)
}

func Struct(dst, src interface{}, opts ...Option) error { return Copy(dst, src, opts...) }

// MapStruct
// map<->struct
// map和结构体相互转化
func MapStruct(dst interface{}, src interface{}, opts ...func(cfg *mapstructure.DecoderConfig)) error {
	var cfg = &mapstructure.DecoderConfig{
		TagName:          "json",
		Metadata:         nil,
		Result:           &dst,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}

	for i := range opts {
		opts[i](cfg)
	}

	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return err
	}

	return decoder.Decode(src)
}
