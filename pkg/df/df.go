package df

import (
	"fmt"
	"reflect"
	"time"

	"github.com/antonmedv/expr/compiler"
	"github.com/antonmedv/expr/parser"
	"github.com/antonmedv/expr/vm"
	"github.com/pubgo/xerror"
)

//func parse(data reflect.Value) []Value {
//	switch data.Kind() {
//	case reflect.Slice, reflect.Array:
//	case reflect.Map:
//	case reflect.Struct:
//	case reflect.Int, reflect.Int8:
//	case reflect.Float32:
//	case reflect.String:
//	case reflect.Ptr:
//	case reflect.Bool:
//	case reflect.Interface:
//	}
//	var dataValue []Value
//	return dataValue
//}

func NewDataFrame(data []interface{}) *DataFrame {
	//var rv = reflect.ValueOf(data)
	//switch rv.Kind() {
	//case reflect.Array, reflect.Slice:
	//	l := rv.Len()
	//	for i := 0; i < l; i++ {
	//		rv.Index(i)
	//	}
	//default:
	//	panic(xerror.Fmt("data type kind(%s) error", rt.Kind()))
	//}

	return &DataFrame{data: data}
}

type DataFrame struct {
	data []interface{}
}

func (t *DataFrame) Map(fn interface{}) *DataFrame {
	var data []interface{}
	var rr = reflect.ValueOf(fn)
	for i := range t.data {
		var vv = rr.Call([]reflect.Value{reflect.ValueOf(t.data[i])})
		data = append(data, vv[0].Interface())
	}
	return &DataFrame{data: data}
}

func (t *DataFrame) Filter(code string) *DataFrame {
	var filterCode = fmt.Sprintf("filter(data, {%s})", code)
	var v = newVm(filterCode)
	return &DataFrame{data: v(t.data).([]interface{})}
}

func (t *DataFrame) Sum(code string) float64 {
	var sum = float64(0)
	var v = newVm(code)
	for i := range t.data {
		output := v(t.data[i])
		switch ret := output.(type) {
		case float64:
			sum += ret
		case float32:
			sum += float64(ret)
		case int:
			sum += float64(ret)
		case int8:
			sum += float64(ret)
		case int16:
			sum += float64(ret)
		case int32:
			sum += float64(ret)
		case int64:
			sum += float64(ret)
		}
	}
	return sum
}

func (t *DataFrame) Avg(code string) float64 {
	if len(t.data) == 0 {
		return 0
	}

	var v = newVm(code)
	var sum = float64(0)
	for i := range t.data {
		output := v(t.data[i])
		switch ret := output.(type) {
		case float64:
			sum += ret
		case float32:
			sum += float64(ret)
		case int:
			sum += float64(ret)
		case int8:
			sum += float64(ret)
		case int16:
			sum += float64(ret)
		case int32:
			sum += float64(ret)
		case int64:
			sum += float64(ret)
		}
	}
	return sum / float64(len(t.data))
}

func (t *DataFrame) Flatten(code string) map[interface{}]interface{} {
	var data = make(map[interface{}]interface{})
	var v = newVm(code)
	for i := range t.data {
		data[v(t.data[i])] = t.data[i]
	}
	return data
}

func newVm(code string) func(data interface{}) interface{} {
	tree, err := parser.Parse(code)
	xerror.Panic(err)

	program, err := compiler.Compile(tree, nil)
	xerror.Panic(err)
	xerror.Assert(program == nil, "program is nil")
	var v vm.VM
	return func(data interface{}) interface{} {
		return xerror.PanicErr(v.Run(program, data))
	}
}

func (t *DataFrame) Reduce(fn func(a, b interface{}) interface{}) interface{} {
	var data interface{}
	for i := range t.data {
		if data == nil {
			data = t.data[i]
			continue
		}

		data = fn(data, t.data[i])
	}
	return data
}

func Tumble(interval int) func(dt interface{ Time() time.Time }) []interface{} {
	var dataFrame = make(map[int][]interface{})
	var latestWindow int
	return func(dt interface{ Time() time.Time }) []interface{} {
		var window = dt.Time().Second() - dt.Time().Second()%interval
		dataFrame[window] = append(dataFrame[window], dt)
		if window != latestWindow {
			var data = dataFrame[latestWindow]
			delete(dataFrame, latestWindow)
			latestWindow = window
			return data
		}
		return nil
	}
}

func (t *DataFrame) GroupBy(code string) map[interface{}]*DataFrame {
	var data = make(map[interface{}]*DataFrame)
	var v = newVm(code)
	for i := range t.data {
		output := v(t.data[i])
		if data[output] == nil {
			data[output] = new(DataFrame)
		}
		data[output].data = append(data[output].data, t.data[i])
	}
	return data
}
