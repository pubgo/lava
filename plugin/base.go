package plugin

import (
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/lug/watcher"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var _ = Plugin(&Base{})
var logs = xlog.Named(Name)

type Base struct {
	Name       string
	OnInit     func(ent interface{})
	OnWatch    func(name string, resp *watcher.Response)
	OnCommands func(cmd *cobra.Command)
	OnFlags    func(flags *pflag.FlagSet)
	OnVars     func(w func(name string, data func() interface{}))
	OnLog      func(log xlog.Xlog)
}

func (p *Base) String() string { return p.Name }
func (p *Base) Init(ent interface{}) (err error) {
	defer xerror.RespErr(&err)

	if p.OnInit != nil {
		logs.Infof("plugin [%s] init", p.Name)
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
	defer xerror.RespExit(Name, "Watch")

	if p.OnWatch == nil {
		return nil
	}

	xlog.Infof("plugin [%s] watch", p.Name)
	p.OnWatch(name, r)

	return
}

func (p *Base) Commands() *cobra.Command {
	defer xerror.RespExit(Name, "Commands")

	if p.OnCommands == nil {
		return nil
	}

	cmd := &cobra.Command{Use: p.Name}
	p.OnCommands(cmd)
	return cmd
}

func (p *Base) Flags() *pflag.FlagSet {
	defer xerror.RespExit(Name, "Flags")

	flags := pflag.NewFlagSet(p.Name, pflag.PanicOnError)
	if p.OnFlags != nil {
		p.OnFlags(flags)
	}
	return flags
}
