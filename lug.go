package lug

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/entry/ctl"
	"github.com/pubgo/lug/entry/grpc"
	"github.com/pubgo/lug/entry/rest"
	"github.com/pubgo/lug/entry/task"
	"github.com/pubgo/lug/internal/runtime"
	"github.com/pubgo/lug/plugin"
)

func BeforeStart(fn func()) { entry.BeforeStart(fn) }
func AfterStart(fn func())  { entry.AfterStart(fn) }
func BeforeStop(fn func())  { entry.BeforeStop(fn) }
func AfterStop(fn func())   { entry.AfterStop(fn) }

func NewTask(name string) task.Entry          { return task.New(name) }
func NewRest(name string) rest.Entry          { return rest.New(name) }
func NewGrpc(name string) grpc.Entry          { return grpc.New(name) }
func NewCtl(name string) ctl.Entry            { return ctl.New(name) }
func Entry(fn func() entry.Entry) entry.Entry { return fn() }

func GetCfg() config.Config                          { return config.GetCfg() }
func Run(short string, entries ...entry.Entry) error { return runtime.Run(short, entries...) }
func Start(ent entry.Entry) error                    { return runtime.Start(ent) }
func Stop(ent entry.Entry) error                     { return runtime.Stop(ent) }

func Plugin(plg plugin.Plugin, opts ...plugin.Opt) { plugin.Register(plg, opts...) }
