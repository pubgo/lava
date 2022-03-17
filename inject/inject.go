package inject

import (
	"fmt"
	"reflect"
	"strings"

	exp "github.com/antonmedv/expr"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/reflectx"
)

var injectHandlers = make(map[string]func(obj Object, field Field) (interface{}, bool))

func Register(typ interface{}, fn func(obj Object, field Field) (interface{}, bool)) {
	if typ == nil {
		panic("[typ] is nil")
	}

	if fn == nil {
		panic("[fn] is nil")
	}

	var typStr = reflect.TypeOf(typ).String()

	t := reflect.TypeOf(typ)
	for t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Interface {
		typStr = t.Elem().String()
	}

	injectHandlers[typStr] = fn
}

func Inject(val interface{}) {
	if val == nil {
		panic("[val] is nil")
	}

	var v reflect.Value
	switch val.(type) {
	case reflect.Value:
		v = val.(reflect.Value)
	default:
		v = reflect.ValueOf(val)
	}
	v = reflectx.Indirect(v)

	if !v.CanSet() {
		panic("[val] should be ptr or interface")
	}

	var obj = Object{Value: v}
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanSet() {
			continue
		}

		var fieldT = v.Type().Field(i)
		var fn = injectHandlers[fieldT.Type.String()]
		if fn == nil {
			continue
		}

		var field = Field{Field: fieldT, val: val}
		var ret, ok = fn(obj, field)
		if !ok {
			continue
		}

		if ret == nil {
			panic("[ret] is nil")
		}

		v.Field(i).Set(reflect.ValueOf(ret))
	}
}

type Object struct {
	Value reflect.Value
}

func (o Object) Name() string {
	return o.Value.Type().Name()
}

type Field struct {
	Field reflect.StructField
	val   interface{}
}

func (f Field) Tag(name string) string {
	return strings.TrimSpace(f.Field.Tag.Get(name))
}

func (f Field) Name() string {
	var name = f.Tag("name")
	if name != "" {
		return name
	}

	var expr = f.Tag("inject-expr")
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
