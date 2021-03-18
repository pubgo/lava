package gutils

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/imdario/mergo"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"github.com/pubgo/x/xutil"
	"github.com/pubgo/xerror"
	"golang.org/x/crypto/scrypt"
)

type nullTransformer struct {
}

func (t nullTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	return func(dst, src reflect.Value) error {
		if src.IsZero() {
			return nil
		}

		if dst.CanSet() {
			dst.Set(src)
		}
		return nil
	}
}

func Mergo(dst, src interface{}, opts ...func(*mergo.Config)) error {
	opts = append(opts, mergo.WithOverride, mergo.WithTypeCheck)
	return xerror.WrapF(mergo.Merge(dst, src, opts...), "\ndst: %#v\n\nsrc: %#v", dst, src)
}

func Map(dst, src interface{}, opts ...func(*mergo.Config)) error {
	opts = append(opts, mergo.WithOverride, mergo.WithTypeCheck)
	return xerror.WrapF(mergo.Map(dst, src, opts...), "\ndst: %#v\n\nsrc: %#v", dst, src)
}

func StructMerge(dst, src interface{}) error {
	var dt = make(map[string]interface{})
	xerror.Exit(Map(&dt, src))

	var opts []func(*mergo.Config)
	opts = append(opts, mergo.WithOverride, mergo.WithTypeCheck)
	opts = append(opts, mergo.WithTransformers(nullTransformer{}))
	return xerror.WrapF(mergo.Map(dst, dt, opts...), "\ndst: %#v\n\nsrc: %#v", dst, src)
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

	xerror.Assert(t == nil, "[t] should not be nil")
	xerror.Assert(fn == nil, "[fn] should not be nil")

	vfn := reflect.ValueOf(fn)
	xerror.Assert(vfn.Type().Kind() != reflect.Func, "[fn] type error, type:%#v", vfn)
	xerror.Assert(vfn.Type().NumIn() != 1, "[fn] input num should be one, now:%d", vfn.Type().NumIn())

	tfn := reflect.TypeOf(t)
	if !tfn.Implements(vfn.Type().In(0)) {
		return nil
	}

	vfn.Call([]reflect.Value{reflect.ValueOf(t)})
	return nil
}

// ParseBool true
func ParseBool(data string) bool {
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
		err = xutil.Try(func() { xerror.Panic(fn()) })
		if err == nil {
			break
		}

		time.Sleep(time.Second)
	}
	return
}
