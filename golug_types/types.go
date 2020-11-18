package golug_types

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/xerror"
)

type CfgValue map[string]interface{}

func (t CfgValue) Decode(fn interface{}) (err error) {
	defer xerror.RespErr(&err)

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "json",
		Metadata:         nil,
		Result:           fn,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	})
	xerror.Panic(err)

	return xerror.Wrap(decoder.Decode(t))
}
