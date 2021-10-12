package lava

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/internal/runtime"
	"github.com/pubgo/lava/plugin"
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
