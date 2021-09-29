package gutil

import (
	_ "unsafe"

	_ "github.com/gin-gonic/gin/binding"
)

//go:linkname MapFormByTag github.com/gin-gonic/gin/binding.mapFormByTag
func MapFormByTag(ptr interface{}, form map[string][]string, tag string) error
