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
)

func Init() (err error)                                          { return golug_config.Init() }
func Start(ent golug_entry.Entry) error                          { return golug_cmd.Start(ent) }
func Stop(ent golug_entry.Entry) error                           { return golug_cmd.Stop(ent) }
func Run(entries ...golug_entry.Entry) (err error)               { return golug_cmd.Run(entries...) }
func NewHttpEntry(name string) golug_entry.HttpEntry             { return http_entry.New(name) }
func NewGrpcEntry(name string) golug_entry.GrpcEntry             { return grpc_entry.New(name) }
func NewCtlEntry(name string) golug_entry.CtlEntry               { return ctl_entry.New(name) }
func RegisterPlugin(plugin golug_plugin.Plugin, opts ...golug_plugin.ManagerOption) error {
	return golug_plugin.Register(plugin, opts...)
}
func WithBeforeStart(fn func(ctx *dix_run.BeforeStartCtx)) error { return dix_run.WithBeforeStart(fn) }
func WithAfterStart(fn func(ctx *dix_run.AfterStartCtx)) error   { return dix_run.WithAfterStart(fn) }
func WithBeforeStop(fn func(ctx *dix_run.BeforeStopCtx)) error   { return dix_run.WithBeforeStop(fn) }
func WithAfterStop(fn func(ctx *dix_run.AfterStopCtx)) error     { return dix_run.WithAfterStop(fn) }
