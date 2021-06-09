package gutil

import (
	_ "github.com/gin-gonic/gin/binding"
	"github.com/gofiber/fiber/v2/utils"

	"reflect"
	_ "unsafe"
)

//go:linkname MapFormByTag github.com/gin-gonic/gin/binding.mapFormByTag
func MapFormByTag(ptr interface{}, form map[string][]string, tag string) error

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
