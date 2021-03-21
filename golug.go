package golug

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/entry"
	"github.com/pubgo/golug/entry/golug_ctl"
	"github.com/pubgo/golug/entry/golug_grpc"
	"github.com/pubgo/golug/entry/golug_rest"
	"github.com/pubgo/golug/entry/golug_task"
	"github.com/pubgo/golug/internal/golug_cmd"
	_ "github.com/pubgo/golug/internal/golug_log"
	"github.com/pubgo/golug/internal/golug_run"
	_ "github.com/pubgo/golug/metric"
	_ "github.com/pubgo/golug/mux"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/xerror"
)

func BeforeStart(fn func()) { golug_run.BeforeStart(fn) }
func AfterStart(fn func())  { golug_run.AfterStart(fn) }
func BeforeStop(fn func())  { golug_run.BeforeStop(fn) }
func AfterStop(fn func())   { golug_run.AfterStop(fn) }

func NewTask(name string) golug_task.Entry { return golug_task.New(name) }
func NewRest(name string) golug_rest.Entry { return golug_rest.New(name) }
func NewGrpc(name string) golug_grpc.Entry { return golug_grpc.New(name) }
func NewCtl(name string) golug_ctl.Entry   { return golug_ctl.New(name) }
func GetCfg() *config.Config               { return config.GetCfg() }
func OnCfg(fn func(cfg *config.Config))    { config.On(fn) }

func Run(entries ...entry.Entry) {
	defer xerror.RespExit()
	xerror.Panic(golug_cmd.Run(entries...))
}

func Plugin(plg plugin.Plugin, opts ...plugin.ManagerOpt) {
	defer xerror.RespExit()
	plugin.Register(plg, opts...)
}
