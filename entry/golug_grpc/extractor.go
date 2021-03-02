package golug_grpc

import (
	"fmt"
	"reflect"
	"strings"

	registry "github.com/pubgo/golug/registry"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func extractValue(v reflect.Type) *registry.Value {
	defer xerror.RespExit("extractValue")

	if v == nil {
		return nil
	}

	arg := &registry.Value{
		Name: v.Name(),
		Type: v.Name(),
	}

	switch v.Kind() {
	case reflect.Ptr:
		v = v.Elem()
		arg.Name = v.Name()
		arg.Type = v.Name()

		switch v.Kind() {
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				f := v.Field(i)
				val := extractValue(f.Type)
				if val == nil {
					continue
				}

				// if we can find a json tag use it
				if tags := f.Tag.Get("json"); len(tags) > 0 {
					parts := strings.Split(tags, ",")
					if parts[0] == "-" || parts[0] == "omitempty" {
						continue
					}
					val.Name = parts[0]
				}

				// if there's no name default it
				if len(val.Name) == 0 {
					val.Name = v.Field(i).Name
				}

				arg.Values = append(arg.Values, val)
			}
		case reflect.Slice:
			p := v.Elem()
			if p.Kind() == reflect.Ptr {
				p = p.Elem()
			}
			arg.Type = "[]" + p.Name()
			val := extractValue(v.Elem())
			if val != nil {
				arg.Values = append(arg.Values, val)
			}
		}
	case reflect.Interface:
		if m, ok := v.MethodByName("SendAndClose"); ok {
			arg.Values = append(arg.Values, extractValue(m.Type.In(0)))
		}

		if m, ok := v.MethodByName("Send"); ok {
			arg.Values = append(arg.Values, extractValue(m.Type.In(0)))
		}

		if m, ok := v.MethodByName("Recv"); ok {
			arg.Values = append(arg.Values, extractValue(m.Type.Out(0)))
		}
	}

	return arg
}

func extractEndpoint(method reflect.Method) *registry.Endpoint {
	defer xerror.RespExit("extractEndpoint")

	if method.PkgPath != "" {
		return nil
	}

	var rspType, reqType reflect.Type
	mt := method.Type

	var reqStream bool
	var respStream bool
	switch mt.NumOut() {
	case 1:
		switch mt.NumIn() {
		case 2:
			reqStream = true
			reqType = mt.In(1)
			rspType = mt.In(1)
			if _, ok := reqType.MethodByName("SendAndClose"); !ok {
				respStream = true
			}
		case 3:
			reqType = mt.In(1)
			rspType = mt.In(2)
			respStream = true
		}
	case 2:
		reqType = mt.In(2)
		rspType = mt.Out(0)
	}

	if rspType == nil {
		xlog.Error("[rspType] is nil")
		return nil
	}

	request := extractValue(reqType)
	response := extractValue(rspType)

	return &registry.Endpoint{
		Name:     method.Name,
		Request:  request,
		Response: response,
		Metadata: map[string]string{
			"req_stream":  fmt.Sprintf("%v", reqStream),
			"resp_stream": fmt.Sprintf("%v", respStream),
		},
	}
}
