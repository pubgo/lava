package golug_util

import (
	"encoding/json"
	"os"

	"github.com/pubgo/xerror"
)

func PathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func MarshalIndent(v interface{}) string {
	return string(xerror.PanicBytes(json.MarshalIndent(v, "", "  ")))
}

func Marshal(v interface{}) string {
	return string(xerror.PanicBytes(json.Marshal(v)))
}
