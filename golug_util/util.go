package golug_util

import (
	"encoding/json"
	"fmt"
	"github.com/pubgo/xerror"
	"os"
	"reflect"
)

// JsonDiff
func JsonDiff(src, target string, depth uint) []string {
	if src == target || reflect.DeepEqual(src, target) {
		return nil
	}

	var srcObj map[string]interface{}
	var targetObj map[string]interface{}
	xerror.Panic(json.Unmarshal([]byte(src), &srcObj))
	xerror.Panic(json.Unmarshal([]byte(target), &targetObj))
	srcKV := Json2KV(srcObj, depth)
	targetKV := Json2KV(targetObj, depth)
	for k, v := range targetKV {
		if dt, ok := srcKV[k]; ok && reflect.DeepEqual(dt, v) {
			delete(targetKV, k)
		}
	}

	var ret []string
	for k := range targetKV {
		ret = append(ret, k)
	}
	return ret
}

func Json2KV(obj interface{}, depth uint) map[string]interface{} {
	var data = make(map[string]interface{})
	rObj := reflect.ValueOf(obj)

	if depth == 0 {
		return map[string]interface{}{"": rObj.Interface()}
	}

	switch rObj.Kind() {
	case reflect.Map:
		iter := rObj.MapRange()
		for iter.Next() {
			for k, v := range Json2KV(iter.Value().Interface(), depth-1) {
				key := iter.Key().String()
				if k != "" {
					key = key + "/" + k
				}
				data[key] = v
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < rObj.Len(); i++ {
			for k, v := range Json2KV(rObj.Index(i).Interface(), depth-1) {
				data[fmt.Sprintf("[%d]/%s", i, k)] = v
			}
		}
	default:
		return map[string]interface{}{"": rObj.Interface()}
	}
	return data
}

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
