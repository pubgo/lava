package inject

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	exp "github.com/antonmedv/expr"
	"github.com/hetiansu5/urlquery"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/reflectx"
)

const (
	injectKey = "inject"
	nameKey   = "name"
)

var typeProviders = make(map[reflect.Type]func(obj Object, field Field) (interface{}, bool))

func WithVal(val interface{}) func(obj Object, field Field) (interface{}, bool) {
	if val == nil {
		panic("[val] is nil")
	}

	return func(obj Object, field Field) (interface{}, bool) { return val, true }
}

func Register(typ interface{}, fn func(obj Object, field Field) (interface{}, bool)) {
	xerror.Assert(typ == nil, "[typ] is nil")
	xerror.Assert(fn == nil, "[fn] is nil")

	t := reflect.TypeOf(typ)
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Interface {
		t = t.Elem()
	}

	typeProviders[t] = fn
}

func Inject(val interface{}) interface{} {
	var v reflect.Value
	switch val.(type) {
	case nil:
		panic("[val] is nil")
	case reflect.Value:
		v = val.(reflect.Value)
	default:
		v = reflect.ValueOf(val)
	}

	v = reflectx.Indirect(v)

	if !v.CanSet() {
		panic(fmt.Sprintf("[val] should be ptr or interface, val=%#v", val))
	}

	var obj = objectImpl{value: v}
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanSet() {
			continue
		}

		var fieldT = v.Type().Field(i)
		var field = fieldImpl{field: fieldT, val: val}
		if tagVal := field.Tag(injectKey); tagVal != "" {
			tagVal = os.Expand(tagVal, func(s string) string {
				if !strings.HasPrefix(s, ".") {
					return os.Getenv(s)
				}

				out, err := exp.Eval(strings.Trim(s, "."), val)
				xerror.Panic(err)
				return fmt.Sprintf("%v", out)
			})
			xerror.Panic(urlquery.Unmarshal([]byte(tagVal), &field.tagVal))
		}

		var fn = typeProviders[fieldT.Type]
		if fn == nil {
			xerror.Assert(field.tagVal.Required, "type(%s) has not provider", fieldT.Type.String())
			continue
		}

		var ret, ok = fn(&obj, &field)
		if !ok {
			continue
		}

		xerror.Assert(ret == nil, "type(%s) provider value is nil", fieldT.Type.String())
		v.Field(i).Set(reflect.ValueOf(ret))
	}
	return val
}

type injectTag struct {
	Name     string `query:"name"`
	Required bool   `query:"required"`
	Expr     bool   `query:"expr"`
}

type objectImpl struct {
	value reflect.Value
}

func (o *objectImpl) Name() string {
	return o.value.Type().Name()
}

func (o *objectImpl) Type() string {
	return o.value.Type().String()
}

type fieldImpl struct {
	tagVal injectTag
	field  reflect.StructField
	val    interface{}
}

func (f fieldImpl) Tag(name string) string {
	return strings.TrimSpace(f.field.Tag.Get(name))
}

func (f fieldImpl) Type() string {
	return f.field.Type.String()
}

func (f fieldImpl) Name() string {
	if f.tagVal.Name != "" {
		return f.tagVal.Name
	}

	var name = f.Tag(nameKey)
	if name != "" {
		return name
	}

	return consts.KeyDefault
}
