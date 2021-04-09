package lug

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/entry/ctlEntry"
	"github.com/pubgo/lug/entry/grpcEntry"
	"github.com/pubgo/lug/entry/restEntry"
	"github.com/pubgo/lug/entry/taskEntry"
	"github.com/pubgo/lug/internal/golug_cmd"
	_ "github.com/pubgo/lug/internal/golug_log"
	"github.com/pubgo/lug/internal/golug_run"
	_ "github.com/pubgo/lug/metric"
	_ "github.com/pubgo/lug/mux"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/xerror"
)

func BeforeStart(fn func()) { golug_run.BeforeStart(fn) }
func AfterStart(fn func())  { golug_run.AfterStart(fn) }
func BeforeStop(fn func())  { golug_run.BeforeStop(fn) }
func AfterStop(fn func())   { golug_run.AfterStop(fn) }

func NewTask(name string) taskEntry.Entry { return taskEntry.New(name) }
func NewRest(name string) restEntry.Entry { return restEntry.New(name) }
func NewGrpc(name string) grpcEntry.Entry { return grpcEntry.New(name) }
func NewCtl(name string) ctlEntry.Entry    { return ctlEntry.New(name) }
func GetCfg() *config.Config               { return config.GetCfg() }
func CfgOn(fn func(cfg *config.Config))    { config.On(fn) }

func Run(entries ...entry.Entry) {
	defer xerror.RespExit()
	xerror.Panic(golug_cmd.Run(entries...))
}

func Plugin(plg plugin.Plugin, opts ...plugin.ManagerOpt) {
	defer xerror.RespExit()
	plugin.Register(plg, opts...)
}
