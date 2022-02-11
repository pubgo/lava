package utils

import (
	"bytes"

	"github.com/bradleyjkemp/memviz"
)

// Memviz 对象内存转化为graphviz
func Memviz(is ...interface{}) []byte {
	var data = bytes.NewBuffer(nil)
	memviz.Map(data, is...)
	return data.Bytes()
}
