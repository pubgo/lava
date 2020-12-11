package golug

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_cmd"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/ctl_entry"
	"github.com/pubgo/golug/golug_entry/grpc_entry"
	"github.com/pubgo/golug/golug_entry/http_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
)

func Init() {
	defer xerror.RespExit()
	xerror.Next().Panic(golug_config.Init())
}
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
func NewHttpEntry(name string) http_entry.HttpEntry { return http_entry.New(name) }
func NewGrpcEntry(name string) grpc_entry.GrpcEntry { return grpc_entry.New(name) }
func NewCtlEntry(name string) ctl_entry.CtlEntry    { return ctl_entry.New(name) }
func RegisterPlugin(plugin golug_plugin.Plugin, opts ...golug_plugin.ManagerOption) {
	defer xerror.RespExit()
	xerror.Next().Panic(golug_plugin.Register(plugin, opts...))
}
func WithBeforeStart(fn func(ctx *dix_run.BeforeStartCtx)) {
	defer xerror.RespExit()
	xerror.Next().Panic(dix_run.WithBeforeStart(fn))
}
func WithAfterStart(fn func(ctx *dix_run.AfterStartCtx)) {
	defer xerror.RespExit()
	xerror.Next().Panic(dix_run.WithAfterStart(fn))
}
func WithBeforeStop(fn func(ctx *dix_run.BeforeStopCtx)) {
	defer xerror.RespExit()
	xerror.Next().Panic(dix_run.WithBeforeStop(fn))
}
func WithAfterStop(fn func(ctx *dix_run.AfterStopCtx)) {
	defer xerror.RespExit()
	xerror.Next().Panic(dix_run.WithAfterStop(fn))
}
