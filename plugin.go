package golug

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
)

func RegisterPlugin(plugin golug_plugin.Plugin, opts ...golug_plugin.ManagerOption) error {
	return xerror.Wrap(golug_plugin.Register(plugin, opts...))
}

func WithBeforeStart(fn func(ctx *dix_run.BeforeStartCtx)) error {
	return xerror.Wrap(dix_run.WithBeforeStart(fn))
}
func WithAfterStart(fn func(ctx *dix_run.AfterStartCtx)) error {
	return xerror.Wrap(dix_run.WithAfterStart(fn))
}
func WithBeforeStop(fn func(ctx *dix_run.BeforeStopCtx)) error {
	return xerror.Wrap(dix_run.WithBeforeStop(fn))
}
func WithAfterStop(fn func(ctx *dix_run.AfterStopCtx)) error {
	return xerror.Wrap(dix_run.WithAfterStop(fn))
}
