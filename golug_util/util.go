package golug_util

import (
	"encoding/json"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/xerror"
)

func PathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func MarshalIndent(v interface{}) string {
	return string(xerror.PanicBytes(json.MarshalIndent(v, "", "  ")))
}

func Marshal(v interface{}) string {
	return string(xerror.PanicBytes(json.Marshal(v)))
}

func Decode(val map[string]interface{}, fn interface{}) (gErr error) {
	defer xerror.RespErr(&gErr)

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

	return xerror.Wrap(decoder.Decode(val))
}
