package merge

import (
	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
)

type Option func(opts *copier.Option)

// Copy
// struct<->struct
// 各种类型结构体之间的field copy
func Copy[A any, B any](dst A, src B, opts ...Option) result.Result[A] {
	var optList = copier.Option{DeepCopy: true, IgnoreEmpty: true}
	for i := range opts {
		opts[i](&optList)
	}

	return result.Wrap(dst, copier.CopyWithOption(dst, src, optList)).OnErr(func(err result.Error) result.Error {
		return err.WithMeta("dst", dst).WithMeta("src", src)
	})
}

func Struct[A any, B any](dst A, src B, opts ...Option) result.Result[A] {
	return Copy(dst, src, opts...)
}

// MapStruct
// map<->struct
// map和结构体相互转化
func MapStruct[A any, B any](dst A, src B, opts ...func(cfg *mapstructure.DecoderConfig)) (r result.Result[A]) {
	defer recovery.Result(&r)

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
	return result.Wrap(dst, decoder.Decode(src)).OnErr(func(err result.Error) result.Error {
		return err.WithMeta("dst", dst).WithMeta("src", src)
	})
}
