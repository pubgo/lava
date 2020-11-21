package golug_entry

import (
	"reflect"

	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type RunEntry interface {
	Init() error
	Start() error
	Stop() error
	Options() Options
}

type Entry interface {
	Description(description ...string) error
	Version(v string) error
	Flags(fn func(flags *pflag.FlagSet)) error
	Commands(commands ...*cobra.Command) error
	Run() RunEntry
	UnWrap(fn interface{}) error
}

type Option func(o *Options)
type Options struct {
	Initialized bool
	Addr        string
	Name        string
	Version     string
	RunCommand  *cobra.Command
	Command     *cobra.Command
}

func UnWrap(t interface{}, fn interface{}) (err error) {
	defer xerror.RespErr(&err)

	if t == nil {
		return xerror.New("[t] should not be nil")
	}

	if fn == nil {
		return xerror.New("[fn] should not be nil")
	}

	_fn := reflect.ValueOf(fn)
	if _fn.Type().Kind() != reflect.Func {
		return xerror.Fmt("[fn] type error, type:%#v", fn)
	}

	if _fn.Type().NumIn() != 1 {
		return xerror.Fmt("[fn] input num should be one, now:%d", _fn.Type().NumIn())
	}

	_t := reflect.TypeOf(t)
	if !_t.Implements(_fn.Type().In(0)) {
		return nil
	}

	_fn.Call([]reflect.Value{reflect.ValueOf(t)})
	return nil
}
