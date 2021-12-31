package debug

import (
	"github.com/pubgo/x/q"
)

func Debug(v ...interface{})         { q.Q(v...) }
func Sdebug(v ...interface{}) []byte { return q.Sq(v...) }
