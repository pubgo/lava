package merge

import (
	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/xerr"
)

type Option func(opts *copier.Option)

// Copy
// struct<->struct
// 各种类型结构体之间的field copy
func Copy(dst interface{}, src interface{}, opts ...Option) error {
	var optList = copier.Option{DeepCopy: true, IgnoreEmpty: true}
	for i := range opts {
		opts[i](&optList)
	}

	return xerr.WrapF(copier.CopyWithOption(dst, src, optList), "\ndst: %#v\n\nsrc: %#v", dst, src)
}

func Struct(dst, src interface{}, opts ...Option) error { return Copy(dst, src, opts...) }

// MapStruct
// map<->struct
// map和结构体相互转化
func MapStruct(dst interface{}, src interface{}, opts ...func(cfg *mapstructure.DecoderConfig)) (err error) {
	defer recovery.Err(&err)

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

	decoder := assert.Must1(mapstructure.NewDecoder(cfg))
	return xerr.Wrap(decoder.Decode(src))
}
