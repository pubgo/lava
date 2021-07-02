package plugin

import (
	"github.com/pubgo/lug/pkg/logutil"
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/lug/watcher"
	"github.com/pubgo/x/try"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var _ = Plugin(&Base{})

type Base struct {
	Name       string
	OnInit     func(ent interface{})
	OnWatch    func(name string, resp *watcher.Response)
	OnCodec    func(name string, resp *watcher.Response) (map[string]interface{}, error)
	OnCommands func(cmd *cobra.Command)
	OnFlags    func(flags *pflag.FlagSet)
	OnVars     func(w func(name string, data func() interface{}))
}

func (p *Base) String() string { return p.Name }
func (p *Base) Init(ent interface{}) (err error) {
	return try.Try(func() {
		if p.OnInit != nil {
			xlog.Infof("plugin [%s] init", p.Name)
			p.OnInit(ent)
		}

		if p.OnVars != nil {
			p.OnVars(vars.Watch)
		}
	})
}

func (p *Base) Codec(name string, resp *watcher.Response) (dt map[string]interface{}, err error) {
	return dt, try.Try(func() {
		if p.OnCodec == nil {
			return
		}

		dt, err = p.OnCodec(name, resp)
		xerror.Panic(err)
	})
}

func (p *Base) Watch(name string, r *watcher.Response) (err error) {
	return try.Try(func() {
		if p.OnWatch == nil {
			return
		}

		xlog.Infof("plugin [%s] watch", p.Name)
		p.OnWatch(name, r)
	})
}

func (p *Base) Commands() *cobra.Command {
	if p.OnCommands == nil {
		return nil
	}

	cmd := &cobra.Command{Use: p.Name}

	logutil.Exit(func() { p.OnCommands(cmd) })

	return cmd
}

func (p *Base) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet(p.Name, pflag.PanicOnError)
	if p.OnFlags != nil {
		logutil.Exit(func() { p.OnFlags(flags) })
	}
	return flags
}
