package golug

import (
	"github.com/pubgo/golug/golug_cmd"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/golug_ctl"
	"github.com/pubgo/golug/golug_entry/golug_grpc"
	"github.com/pubgo/golug/golug_entry/golug_rest"
	"github.com/pubgo/golug/golug_entry/golug_task"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/internal/golug_run"
	"github.com/pubgo/xerror"
)

func Run(entries ...golug_entry.Entry) {
	defer xerror.RespExit()
	xerror.Panic(golug_cmd.Run(entries...))
}

func NewTask(name string) golug_task.Entry { return golug_task.New(name) }
func NewRest(name string) golug_rest.Entry { return golug_rest.New(name) }
func NewGrpc(name string) golug_grpc.Entry { return golug_grpc.New(name) }
func NewCtl(name string) golug_ctl.Entry   { return golug_ctl.New(name) }
func RegisterPlugin(plugin golug_plugin.Plugin, opts ...golug_plugin.ManagerOption) {
	defer xerror.RespExit()
	golug_plugin.Register(plugin, opts...)
}

func BeforeStart(fn func()) { golug_run.BeforeStart(fn) }
func AfterStart(fn func())  { golug_run.AfterStart(fn) }
func BeforeStop(fn func())  { golug_run.BeforeStop(fn) }
func AfterStop(fn func())   { golug_run.AfterStop(fn) }
