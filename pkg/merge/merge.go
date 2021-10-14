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
func Copy(dst interface{}, src interface{}, opts ...Option) interface{} {
	var optList copier.Option
	for i := range opts {
		opts[i](&optList)
	}

	xerror.PanicF(copier.CopyWithOption(dst, src, optList), "\ndst: %#v\n\nsrc: %#v", dst, src)
	return dst
}

func Struct(dst, src interface{}, opts ...Option) interface{} { return Copy(dst, src, opts...) }

// MapStruct
// map<->struct
// map和结构体相互转化
func MapStruct(dst interface{}, src interface{}, opts ...func(cfg *mapstructure.DecoderConfig)) interface{} {
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
	xerror.Panic(err)

	xerror.Panic(decoder.Decode(src))
	return dst
}
