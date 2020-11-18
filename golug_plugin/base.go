package golug_plugin

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_types"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var _ Plugin = (*Base)(nil)

type Base struct {
	Name       string
	Enabled    bool `yaml:"enabled" json:"enabled" toml:"enabled"`
	OnInit     func(r golug_types.CfgValue)
	OnWatch    func(r golug_types.CfgValue)
	OnCommands func(cmd *cobra.Command)
	OnHandler  func() fiber.Handler
	OnFlags    func(flags *pflag.FlagSet)
}

func (p *Base) Init(r golug_types.CfgValue) (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(golug_config.Decode(p.Name, p))

	var status = "disabled"
	if p.Enabled {
		status = "enabled"
	}

	xlog.Debugf("plugin [%s] init, status: %s", p.Name, status)

	if !p.Enabled {
		return nil
	}

	if p.OnInit != nil {
		xlog.Debugf("[%s] start init", p.Name)
		p.OnInit(r)
	}
	return nil
}

func (p *Base) Watch(r golug_types.CfgValue) (err error) {
	defer xerror.RespErr(&err)

	if !p.Enabled {
		return nil
	}

	if p.OnWatch != nil {
		xlog.Debugf("[%s] start watch", p.Name)
		p.OnWatch(r)
	}
	return nil
}

func (p *Base) Commands() *cobra.Command {
	if !p.Enabled {
		return nil
	}

	if p.OnCommands != nil {
		cmd := &cobra.Command{Use: p.Name}
		p.OnCommands(cmd)
		return cmd
	}
	return nil
}

func (p *Base) Handler() fiber.Handler {
	if !p.Enabled {
		return nil
	}

	if p.OnHandler != nil {
		return p.OnHandler()
	}
	return nil
}

func (p *Base) String() string {
	return p.Name
}

func (p *Base) Flags() *pflag.FlagSet {
	if !p.Enabled {
		return nil
	}

	flags := pflag.NewFlagSet(p.Name, pflag.PanicOnError)
	if p.OnFlags != nil {
		p.OnFlags(flags)
	}
	return flags
}
