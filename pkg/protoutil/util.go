package protoutil

import (
	"encoding/base64"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/pubgo/x/abc"
	"github.com/pubgo/xerror"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/known/structpb"
)

func NewValue(rv interface{}) abc.Value {
	return abc.NewValue(func() (interface{}, error) {
		switch v := rv.(type) {
		case nil:
			return structpb.NewNullValue(), nil
		case bool:
			return structpb.NewBoolValue(v), nil
		case int:
			return structpb.NewNumberValue(float64(v)), nil
		case int32:
			return structpb.NewNumberValue(float64(v)), nil
		case int64:
			return structpb.NewNumberValue(float64(v)), nil
		case uint:
			return structpb.NewNumberValue(float64(v)), nil
		case uint32:
			return structpb.NewNumberValue(float64(v)), nil
		case uint64:
			return structpb.NewNumberValue(float64(v)), nil
		case float32:
			return structpb.NewNumberValue(float64(v)), nil
		case float64:
			return structpb.NewNumberValue(v), nil
		case string:
			if !utf8.ValidString(v) {
				return nil, xerror.Wrap(protoimpl.X.NewError("invalid UTF-8 in string: %q", v))
			}
			return structpb.NewStringValue(v), nil
		case []byte:
			s := base64.StdEncoding.EncodeToString(v)
			return structpb.NewStringValue(s), nil
		case map[string]interface{}:
			val := NewStruct(v)
			return structpb.NewStructValue(val.Expect("").(*structpb.Struct)), nil
		case []interface{}:
			return NewList(v).Throw(v), nil
		case *structpb.Value:
			if v == nil || v.GetKind() == nil {
				return NewValue(nil).Throw(), nil
			}

			return v, nil
		case structpb.Value:
			return &v, nil
		case *structpb.ListValue:
			return structpb.NewListValue(v), nil
		case structpb.ListValue:
			return structpb.NewListValue(&v), nil
		case *structpb.Struct:
			return structpb.NewStructValue(v), nil
		case structpb.Struct:
			return structpb.NewStructValue(&v), nil
		default:
			switch reflect.TypeOf(v).Kind() {
			case reflect.Map:
				var vv = reflect.ValueOf(v)
				x := &structpb.Struct{Fields: make(map[string]*structpb.Value, vv.Len())}
				var iter = vv.MapRange()
				for iter.Next() {
					k := iter.Key().String()

					if !utf8.ValidString(k) {
						return nil, xerror.Wrap(protoimpl.X.NewError("invalid UTF-8 in string: %q", k))
					}

					var val interface{} = nil
					if v1 := iter.Value(); v1.IsValid() {
						val = v1.Interface()
					}

					x.Fields[k] = NewValue(val).Throw(val).(*structpb.Value)
				}
				return structpb.NewStructValue(x), nil
			case reflect.Struct:
				var vv = reflect.ValueOf(v)
				var tt = vv.Type()
				x := &structpb.Struct{Fields: make(map[string]*structpb.Value, vv.NumField())}
				for i := 0; i < vv.NumField(); i++ {
					var name = tt.Field(i).Name
					if 'a' <= name[0] && name[0] <= 'z' {
						continue
					}

					var val interface{} = nil
					if v1 := vv.Field(i); v1.IsValid() {
						val = v1.Interface()
					}

					if nm := strings.Split(tt.Field(i).Tag.Get("json"), ",")[0]; nm != "" {
						name = nm
					}

					x.Fields[name] = NewValue(val).Throw(val).(*structpb.Value)
				}
				return structpb.NewStructValue(x), nil
			case reflect.Ptr:
				var vv = reflect.ValueOf(v)
				if vv.IsValid() && !vv.IsNil() {
					return NewValue(vv.Elem().Interface()).Throw(), nil
				}
				return NewValue(nil).Throw(), nil
			case reflect.Slice, reflect.Array:
				var vv = reflect.ValueOf(v)
				x := &structpb.ListValue{Values: make([]*structpb.Value, vv.Len())}
				for i := 0; i < vv.Len(); i++ {

					var val interface{} = nil
					if v1 := vv.Index(i); v1.IsValid() && !v1.IsNil() {
						val = v1.Interface()
					}

					x.Values[i] = NewValue(val).Throw().(*structpb.Value)
				}
				return structpb.NewListValue(x), nil
			}

			return nil, xerror.Wrap(protoimpl.X.NewError("invalid type: %T", v))
		}
	})
}

func NewStruct(v map[string]interface{}) abc.Value {
	return abc.NewValue(func() (interface{}, error) {
		x := &structpb.Struct{Fields: make(map[string]*structpb.Value, len(v))}
		for k, v := range v {
			if !utf8.ValidString(k) {
				return nil, xerror.Wrap(protoimpl.X.NewError("invalid UTF-8 in string: %q", k))
			}
			x.Fields[k] = NewValue(v).Throw(v).(*structpb.Value)
		}
		return x, nil
	})
}

func NewList(vv []interface{}) abc.Value {
	return abc.NewValue(func() (interface{}, error) {
		x := &structpb.ListValue{Values: make([]*structpb.Value, len(vv))}
		for i, v := range vv {
			var val = NewValue(v)
			x.Values[i] = val.Throw(v).(*structpb.Value)
		}
		return structpb.NewListValue(x), nil
	})
}

func Zero(rv interface{}) (interface{}, error) {
	switch v := rv.(type) {
	case nil:
		return "null", nil
	case bool:
		return false, nil
	case int,int32,int64,uint,uint32,uint64:
		return 0, nil
	case float32,float64:
		return 0.0, nil
	case string:
		return "", nil
	case []byte:
		return "", nil
	case map[string]interface{}:
		return "{}", nil
	case []interface{}:
		return "[]", nil
	case *structpb.Value:
		return "{}", nil
	case *structpb.ListValue:
		return structpb.NewListValue(v), nil
	case structpb.ListValue:
		return structpb.NewListValue(&v), nil
	case *structpb.Struct:
		return structpb.NewStructValue(v), nil
	case structpb.Struct:
		return structpb.NewStructValue(&v), nil
	default:
		switch reflect.TypeOf(v).Kind() {
		case reflect.Map:
			var vv = reflect.ValueOf(v)
			x := &structpb.Struct{Fields: make(map[string]*structpb.Value, vv.Len())}
			var iter = vv.MapRange()
			for iter.Next() {
				k := iter.Key().String()

				if !utf8.ValidString(k) {
					return nil, xerror.Wrap(protoimpl.X.NewError("invalid UTF-8 in string: %q", k))
				}

				var val interface{} = nil
				if v1 := iter.Value(); v1.IsValid() {
					val = v1.Interface()
				}

				x.Fields[k] = NewValue(val).Throw(val).(*structpb.Value)
			}
			return structpb.NewStructValue(x), nil
		case reflect.Struct:
			var vv = reflect.ValueOf(v)
			var tt = vv.Type()
			x := &structpb.Struct{Fields: make(map[string]*structpb.Value, vv.NumField())}
			for i := 0; i < vv.NumField(); i++ {
				var name = tt.Field(i).Name
				if 'a' <= name[0] && name[0] <= 'z' {
					continue
				}

				var val interface{} = nil
				if v1 := vv.Field(i); v1.IsValid() {
					val = v1.Interface()
				}

				if nm := strings.Split(tt.Field(i).Tag.Get("json"), ",")[0]; nm != "" {
					name = nm
				}

				x.Fields[name] = NewValue(val).Throw(val).(*structpb.Value)
			}
			return structpb.NewStructValue(x), nil
		case reflect.Ptr:
			var vv = reflect.ValueOf(v)
			if vv.IsValid() && !vv.IsNil() {
				return NewValue(vv.Elem().Interface()).Throw(), nil
			}
			return NewValue(nil).Throw(), nil
		case reflect.Slice, reflect.Array:
			var vv = reflect.ValueOf(v)
			x := &structpb.ListValue{Values: make([]*structpb.Value, vv.Len())}
			for i := 0; i < vv.Len(); i++ {

				var val interface{} = nil
				if v1 := vv.Index(i); v1.IsValid() && !v1.IsNil() {
					val = v1.Interface()
				}

				x.Values[i] = NewValue(val).Throw().(*structpb.Value)
			}
			return structpb.NewListValue(x), nil
		}

		return nil, xerror.Wrap(protoimpl.X.NewError("invalid type: %T", v))
	}
}
