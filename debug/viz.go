package debug

import (
	"os"

	"github.com/bradleyjkemp/memviz"
	"github.com/pubgo/xerror"
)

// Memviz 获取对象可视化内存
func Memviz(filename string, is ...interface{}) {
	var f, err = os.Create(filename)
	xerror.Panic(err)
	memviz.Map(f, is...)
}
