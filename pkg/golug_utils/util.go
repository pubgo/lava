package golug_utils

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/imdario/mergo"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/xerror"
	"golang.org/x/crypto/scrypt"
)

func Mergo(dst, src interface{}, opts ...func(*mergo.Config)) {
	opts = append(opts, mergo.WithOverride, mergo.WithTypeCheck)
	xerror.Panic(mergo.Map(dst, src, opts...))
}

func PathExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	xerror.Panic(err)
	return false
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

// IsTrue true
func IsTrue(data string) bool {
	switch strings.ToUpper(data) {
	case "TRUE", "T", "1", "OK", "GOOD", "REAL", "ACTIVE", "ENABLED":
		return true
	default:
		return false
	}
}

func EncodePassword(unencoded string) string {
	newPassword, _ := scrypt.Key([]byte(unencoded), []byte("!#@FDEWREWR&*("), 16384, 8, 1, 64)
	return fmt.Sprintf("%x", newPassword)
}

func Retry(c int, fn func() error) (err error) {
	for i := 0; i < c; i++ {
		if err = fn(); err == nil {
			break
		}

		time.Sleep(time.Second)
	}
	return
}
