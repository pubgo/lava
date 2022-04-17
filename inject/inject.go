package inject

import (
	"fmt"
	"reflect"
	"strings"

	exp "github.com/antonmedv/expr"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/reflectx"
)

const (
	injectKey  = "inject"
	nameKey    = "name"
	injectExpr = "inject-expr"
)

var injectHandlers = make(map[reflect.Type]func(obj Object, field Field) (interface{}, bool))

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

	injectHandlers[t] = fn
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
		var fn = injectHandlers[fieldT.Type]
		if fn == nil {
			continue
		}

		var field = fieldImpl{field: fieldT, val: val}
		var ret, ok = fn(&obj, &field)
		if !ok {
			continue
		}

		if ret == nil {
			panic("[ret] is nil")
		}

		v.Field(i).Set(reflect.ValueOf(ret))
	}
	return val
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
	field reflect.StructField
	val   interface{}
}

func (f fieldImpl) Tag(name string) string {
	return strings.TrimSpace(f.field.Tag.Get(name))
}

func (f fieldImpl) Type() string {
	return f.field.Type.String()
}

func (f fieldImpl) Name() string {
	var name = f.Tag(nameKey)
	if name != "" {
		return name
	}

	var expr = f.Tag(injectExpr)
	if expr != "" {
		out, err := exp.Eval(expr, f.val)
		if err != nil {
			panic(err)
		}

		if out != "" {
			return fmt.Sprintf("%v", out)
		}
	}

	return consts.KeyDefault
}
