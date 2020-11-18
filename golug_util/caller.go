package golug_util

import (
	"github.com/pubgo/xerror"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

func CallerWithFunc(fn interface{}) (string, error) {
	if fn == nil {
		return "", xerror.New("params is nil")
	}

	var _fn = reflect.ValueOf(fn)
	if !_fn.IsValid() || _fn.IsNil() || _fn.Kind() != reflect.Func {
		return "", xerror.New("not func type or type is nil")
	}

	var _e = runtime.FuncForPC(_fn.Pointer())
	var file, line = _e.FileLine(_fn.Pointer())

	var buf = &strings.Builder{}
	defer buf.Reset()

	files := strings.Split(file, string(os.PathSeparator))
	if len(files) > 2 {
		file = filepath.Join(files[len(files)-2:]...)
	}

	buf.WriteString(file)
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(line))
	buf.WriteString(" ")

	ma := strings.Split(_e.Name(), ".")
	buf.WriteString(ma[len(ma)-1])
	return buf.String(), nil
}
