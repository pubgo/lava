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

func Init() {}
func Start(ent golug_entry.Entry) {
	defer xerror.RespExit()
	xerror.Next().Panic(golug_cmd.Start(ent))
}
func Stop(ent golug_entry.Entry) {
	defer xerror.RespExit()
	xerror.Next().Panic(golug_cmd.Stop(ent))
}
func Run(entries ...golug_entry.Entry) {
	defer xerror.RespExit()
	xerror.Next().Panic(golug_cmd.Run(entries...))
}
func NewRestEntry(name string, cfg interface{}) golug_rest.Entry { return golug_rest.New(name, cfg) }
func NewGrpcEntry(name string, cfg interface{}) golug_grpc.Entry { return golug_grpc.New(name, cfg) }
func NewCtlEntry(name string, cfg interface{}) golug_ctl.Entry   { return golug_ctl.New(name, cfg) }
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
