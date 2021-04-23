package plugin

import (
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/lug/watcher"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var _ Plugin = (*Base)(nil)

type Base struct {
	Name       string
	OnInit     func(ent interface{})
	OnWatch    func(name string, resp *watcher.Response)
	OnCommands func(cmd *cobra.Command)
	OnFlags    func(flags *pflag.FlagSet)
	OnVars     func(w func(name string, data func() interface{}))
	OnLog      func(logs xlog.Xlog)
}

func (p *Base) Init(ent interface{}) (err error) {
	defer xerror.RespErr(&err)

	if p.OnInit != nil {
		xlog.Debugf("plugin [%s] init", p.Name)
		p.OnInit(ent)
	}

	if p.OnVars != nil {
		p.OnVars(vars.Watch)
	}

	if p.OnLog != nil {
		xlog.Watch(p.OnLog)
	}

	return nil
}

func (p *Base) Watch(name string, r *watcher.Response) (err error) {
	defer xerror.RespErr(&err)

	if p.OnWatch != nil {
		xlog.Debugf("[%s] start watch", p.Name)
		p.OnWatch(name, r)
	}
	return nil
}

func (p *Base) Commands() *cobra.Command {
	defer xerror.Resp(func(err xerror.XErr) { xlog.Error("command error", zap.Any("err", err)) })

	if p.OnCommands != nil {
		cmd := &cobra.Command{Use: p.Name}
		p.OnCommands(cmd)
		return cmd
	}
	return nil
}

func (p *Base) String() string { return p.Name }

func (p *Base) Flags() *pflag.FlagSet {
	defer xerror.Resp(func(err xerror.XErr) { xlog.Error("flags error", zap.Any("err", err)) })

	flags := pflag.NewFlagSet(p.Name, pflag.PanicOnError)
	if p.OnFlags != nil {
		p.OnFlags(flags)
	}
	return flags
}
