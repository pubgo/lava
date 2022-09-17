package fields

import (
	"database/sql/driver"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
)

var _ binding.Binding = (*bindImpl)(nil)
var Binding = &bindImpl{}

type bindImpl struct {
}

func (f *bindImpl) Name() string { return "field" }

func (f *bindImpl) Bind(request *http.Request, i interface{}) error {
	return bindBuild(i, request.URL.Query())
}

func bindBuild(val interface{}, values url.Values) error {
	var v = reflect.ValueOf(val).Elem()
	var t = v.Type()

	var params = make(map[string]map[string]string)
	for key, vv := range values {
		if len(vv) == 0 {
			continue
		}

		var names = strings.Split(key, "__")
		if len(names) != 2 {
			continue
		}

		if params[names[0]] == nil {
			params[names[0]] = make(map[string]string)
		}

		params[names[0]][names[1]] = vv[0]
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" && !f.Anonymous {
			continue
		}

		if !v.Field(i).CanSet() {
			continue
		}

		var name, ok = f.Tag.Lookup("field")
		if !ok || name == "" {
			continue
		}

		if params[name] == nil {
			continue
		}

		vPtr := v.Field(i)
		if vPtr.IsNil() {
			vPtr = reflect.New(vPtr.Type().Elem())
		}

		if ff, ok := vPtr.Interface().(Field); ok {
			ff.setName(name)
			var data = make(map[string]driver.Valuer)
			for cond, value := range params[name] {
				vv, err := ff.handler(value)
				if err != nil {
					return err
				}
				data[cond] = vv
			}
			ff.setValue(data)
			v.Field(i).Set(vPtr)
		}
	}

	return nil
}
