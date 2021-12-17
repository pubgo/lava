package lavax

import (
	"encoding/json"
	"go/format"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/consts"
)

func EqualFieldType(out interface{}, kind reflect.Kind, key string) bool {
	// Get type of interface
	outTyp := reflect.TypeOf(out).Elem()
	// Must be a struct to match a field
	if outTyp.Kind() != reflect.Struct {
		return false
	}
	// Copy interface to an value to be used
	outVal := reflect.ValueOf(out).Elem()
	// Loop over each field
	for i := 0; i < outTyp.NumField(); i++ {
		// Get field value data
		structField := outVal.Field(i)
		// Can this field be changed?
		if !structField.CanSet() {
			continue
		}
		// Get field key data
		typeField := outTyp.Field(i)
		// Get type of field key
		structFieldKind := structField.Kind()
		// Does the field type equals input?
		if structFieldKind != kind {
			continue
		}
		// Get tag from field if exist
		inputFieldName := typeField.Tag.Get(key)
		if inputFieldName == "" {
			inputFieldName = typeField.Name
		}
		// Compare field/tag with provided key
		if ToLower(inputFieldName) == key {
			return true
		}
	}
	return false
}

func CodeFormat(data ...string) string {
	var str = ""
	for i := range data {
		str += strings.TrimSpace(data[i]) + "\n"
	}
	str = strings.TrimSpace(str)
	return string(xerror.PanicBytes(format.Source([]byte(str))))
}

func DotJoin(str ...string) string {
	return strings.Join(str, ".")
}

// DirExists function to check if directory exists?
func DirExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		// path is a directory
		return true
	}
	return false
}

// FileExists function to check if file exists?
func FileExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		return false
	}
}

func FirstNotEmpty(fx ...func() string) string {
	for i := range fx {
		if s := fx[i](); s != "" {
			return s
		}
	}
	return ""
}

func IfEmpty(str string, fx func()) {
	if str == "" {
		fx()
	}
}

func GetDefault(names ...string) string {
	var name = consts.KeyDefault
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}
	return name
}

func Cost(fn func()) (dur time.Duration, err error) {
	defer func(t time.Time) { dur = time.Since(t) }(time.Now())
	defer xerror.RespErr(&err)
	fn()
	return
}

func TransformObject2Param(object interface{}) (params map[string]string) {
	params = make(map[string]string)
	if object != nil {
		valueOf := reflect.ValueOf(object)
		typeOf := reflect.TypeOf(object)
		if reflect.TypeOf(object).Kind() == reflect.Ptr {
			valueOf = reflect.ValueOf(object).Elem()
			typeOf = reflect.TypeOf(object).Elem()
		}
		numField := valueOf.NumField()
		for i := 0; i < numField; i++ {
			tag := typeOf.Field(i).Tag.Get("param")
			if len(tag) > 0 && tag != "-" {
				switch valueOf.Field(i).Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16,
					reflect.Int32, reflect.Int64:
					params[tag] = strconv.FormatInt(valueOf.Field(i).Int(), 10)
				case reflect.Uint, reflect.Uint8, reflect.Uint16,
					reflect.Uint32, reflect.Uint64:
					params[tag] = strconv.FormatUint(valueOf.Field(i).Uint(), 10)
				case reflect.Float32, reflect.Float64:
					params[tag] = strconv.FormatFloat(valueOf.Field(i).Float(), 'f', -1, 64)
				case reflect.Bool:
					params[tag] = strconv.FormatBool(valueOf.Field(i).Bool())
				case reflect.String:
					if len(valueOf.Field(i).String()) > 0 {
						params[tag] = valueOf.Field(i).String()
					}
				case reflect.Map:
					if !valueOf.Field(i).IsNil() {
						bytes, err := json.Marshal(valueOf.Field(i).Interface())
						if err != nil {
							panic(err)
						} else {
							params[tag] = string(bytes)
						}
					}
				case reflect.Slice:
					if ss, ok := valueOf.Field(i).Interface().([]string); ok {
						var pv string
						for _, sv := range ss {
							pv += sv + ","
						}
						if strings.HasSuffix(pv, ",") {
							pv = pv[:len(pv)-1]
						}
						if len(pv) > 0 {
							params[tag] = pv
						}
					}
				}
			}
		}
	}
	return
}
