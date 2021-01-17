package golug

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_cmd"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/golug_ctl"
	"github.com/pubgo/golug/golug_entry/golug_grpc"
	"github.com/pubgo/golug/golug_entry/golug_rest"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
)

func Run(entries ...golug_entry.Entry) {
	defer xerror.RespExit()
	xerror.Next().Panic(golug_cmd.Run(entries...))
}

func NewRestEntry(name string) golug_rest.Entry { return golug_rest.New(name) }
func NewGrpcEntry(name string) golug_grpc.Entry { return golug_grpc.New(name) }
func NewCtlEntry(name string) golug_ctl.Entry   { return golug_ctl.New(name) }
func RegisterPlugin(plugin golug_plugin.Plugin, opts ...golug_plugin.ManagerOption) {
	defer xerror.RespExit()
	golug_plugin.Register(plugin, opts...)
}

func BeforeStart(fn func(ctx *dix_run.BeforeStartCtx)) {
	defer xerror.RespExit()
	xerror.Next().Panic(dix_run.WithBeforeStart(fn))
}

func AfterStart(fn func(ctx *dix_run.AfterStartCtx)) {
	defer xerror.RespExit()
	xerror.Next().Panic(dix_run.WithAfterStart(fn))
}

func BeforeStop(fn func(ctx *dix_run.BeforeStopCtx)) {
	defer xerror.RespExit()
	xerror.Next().Panic(dix_run.WithBeforeStop(fn))
}

func AfterStop(fn func(ctx *dix_run.AfterStopCtx)) {
	defer xerror.RespExit()
	xerror.Next().Panic(dix_run.WithAfterStop(fn))
}
