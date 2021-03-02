package plugin

import (
	"github.com/pubgo/golug/watcher"
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
	OnWatch    func(resp *watcher.Response)
	OnCommands func(cmd *cobra.Command)
	OnFlags    func(flags *pflag.FlagSet)
}

func (p *Base) Init(ent interface{}) (err error) {
	defer xerror.RespErr(&err)

	if p.OnInit != nil {
		xlog.Debugf("plugin [%s] init", p.Name)
		p.OnInit(ent)
	}

	return nil
}

func (p *Base) Watch(r *watcher.Response) (err error) {
	defer xerror.RespErr(&err)

	if p.OnWatch != nil {
		xlog.Debugf("[%s] start watch", p.Name)
		p.OnWatch(r)
	}
	return nil
}

func (p *Base) Commands() *cobra.Command {
	defer xerror.Resp(func(err xerror.XErr) { xlog.Error(err.Stack(), zap.Any("err", "command error")) })

	if p.OnCommands != nil {
		cmd := &cobra.Command{Use: p.Name}
		p.OnCommands(cmd)
		return cmd
	}
	return nil
}

func (p *Base) String() string {
	return p.Name
}

func (p *Base) Flags() *pflag.FlagSet {
	defer xerror.Resp(func(err xerror.XErr) { xlog.Error("flags error", zap.Any("err", err)) })

	flags := pflag.NewFlagSet(p.Name, pflag.PanicOnError)
	if p.OnFlags != nil {
		p.OnFlags(flags)
	}
	return flags
}
