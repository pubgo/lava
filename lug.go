package lug

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/internal/runtime"
	"github.com/pubgo/lug/plugin"
)

func BeforeStart(fn func()) { entry.BeforeStart(fn) }
func AfterStart(fn func())  { entry.AfterStart(fn) }
func BeforeStop(fn func())  { entry.BeforeStop(fn) }
func AfterStop(fn func())   { entry.AfterStop(fn) }

func Config() config.Config                          { return config.GetCfg() }
func Run(description string, entries ...entry.Entry) { runtime.Run(description, entries...) }
func Start(ent entry.Entry)                          { runtime.Start(ent) }
func Stop(ent entry.Entry)                           { runtime.Stop(ent) }

func Plugin(plg plugin.Plugin, opts ...plugin.Opt) { plugin.Register(plg, opts...) }
