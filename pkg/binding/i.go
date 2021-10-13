package binding

import (
	_ "unsafe"

	_ "github.com/gin-gonic/gin/binding"
)

//go:linkname MapFormByTag github.com/gin-gonic/gin/binding.mapFormByTag
func MapFormByTag(obj interface{}, form map[string][]string, tag string) error

func MapForm(obj interface{}, form map[string][]string) error { return MapFormByTag(obj, form, "json") }
