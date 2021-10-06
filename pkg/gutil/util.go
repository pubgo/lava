package gutil

import (
	"go/format"
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/pubgo/xerror"
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
		if utils.ToLower(inputFieldName) == key {
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

func ExampleFmt(data ...string) string {
	var str = ""
	for i := range data {
		str += "  " + data[i] + "\n"
	}
	return "  " + strings.TrimSpace(str)
}

func DotJoin(str ...string) string {
	return strings.Join(str, ".")
}

func GetPort(addr string) string {
	var addrList = strings.Split(addr, ":")
	return addrList[len(addrList)-1]
}
