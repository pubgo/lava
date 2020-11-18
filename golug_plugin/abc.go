package golug_plugin

import (
	"encoding/json"

	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Manager interface {
	Plugins(...ManagerOption) []Plugin
	Register(Plugin, ...ManagerOption) error
}

type ManagerOption func(o *ManagerOptions)
type ManagerOptions struct {
	Module string
}

type Response struct {
	Event    string
	Key      []byte
	Value    []byte
	Revision int64
}

func (t *Response) Decode(val interface{}) (err error) {
	defer xerror.RespErr(&err)
	return xerror.Wrap(json.Unmarshal(t.Value, val))
}

type Plugin interface {
	Watch(r *Response) error
	Init(ent golug_entry.Entry) error
	Flags() *pflag.FlagSet
	Commands() *cobra.Command
	String() string
}

type Option func(o *Options)
type Options struct {
	Name     string
	Flags    *pflag.FlagSet
	Commands *cobra.Command
}
