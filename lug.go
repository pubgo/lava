package lug

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/entry/ctl"
	"github.com/pubgo/lug/entry/rest"
	"github.com/pubgo/lug/entry/rpc"
	"github.com/pubgo/lug/entry/task"
	_ "github.com/pubgo/lug/internal/log_plugin"
	"github.com/pubgo/lug/internal/runtime"
	_ "github.com/pubgo/lug/metric"
	_ "github.com/pubgo/lug/mux"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/xerror"
)

func BeforeStart(fn func()) { runtime.BeforeStart(fn) }
func AfterStart(fn func())  { runtime.AfterStart(fn) }
func BeforeStop(fn func())  { runtime.BeforeStop(fn) }
func AfterStop(fn func())   { runtime.AfterStop(fn) }

func NewTask(name string) task.Entry    { return task.New(name) }
func NewRest(name string) rest.Entry    { return rest.New(name) }
func NewRpc(name string) rpc.Entry      { return rpc.New(name) }
func NewCtl(name string) ctl.Entry      { return ctl.New(name) }
func GetCfg() *config.Config            { return config.GetCfg() }
func CfgOn(fn func(cfg *config.Config)) { config.On(fn) }

func Run(entries ...entry.Abc) {
	defer xerror.RespExit()
	xerror.Panic(runtime.Run(entries...))
}

func Start(ent entry.Abc) error { return runtime.Start(ent) }
func Stop(ent entry.Abc) error  { return runtime.Stop(ent) }

func Plugin(plg plugin.Plugin, opts ...plugin.ManagerOpt) {
	defer xerror.RespExit()
	plugin.Register(plg, opts...)
}
