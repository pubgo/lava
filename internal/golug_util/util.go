package golug_util

import (
	"os"
	"reflect"

	jsoniter "github.com/json-iterator/go"
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
	return string(xerror.PanicBytes(jsoniter.MarshalIndent(v, "", "  ")))
}

func Marshal(v interface{}) string {
	return string(xerror.PanicBytes(jsoniter.Marshal(v)))
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

func UnWrap(t interface{}, fn interface{}) (err error) {
	defer xerror.RespErr(&err)

	if t == nil {
		return xerror.New("[t] should not be nil")
	}

	if fn == nil {
		return xerror.New("[fn] should not be nil")
	}

	_fn := reflect.ValueOf(fn)
	if _fn.Type().Kind() != reflect.Func {
		return xerror.Fmt("[fn] type error, type:%#v", fn)
	}

	if _fn.Type().NumIn() != 1 {
		return xerror.Fmt("[fn] input num should be one, now:%d", _fn.Type().NumIn())
	}

	_t := reflect.TypeOf(t)
	if !_t.Implements(_fn.Type().In(0)) {
		return nil
	}

	_fn.Call([]reflect.Value{reflect.ValueOf(t)})
	return nil
}
