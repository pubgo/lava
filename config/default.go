package config

import (
	"path/filepath"
	"strings"

	"github.com/pubgo/xerror"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/runenv"
)

var (
	CfgType = "yaml"
	CfgName = "config"
	Home    = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), ".lava")
	CfgPath = ""
	cfg     = &configImpl{v: viper.New()}
)

func init() {
	env.GetWith(&CfgType, "cfg_type", "config_type")
	env.GetWith(&CfgName, "cfg_name", "config_name")
	env.GetWith(&Home, "project_home", "config_home", "config_dir")
}

func Init() error                                  { return getCfg().Init() }
func GetCfg() Config                               { return getCfg() }
func Decode(name string, fn interface{}) error     { return getCfg().Decode(name, fn) }
func GetMap(keys ...string) map[string]interface{} { return getCfg().GetMap(strings.Join(keys, ".")) }

func getCfg() *configImpl {
	xerror.Assert(cfg == nil, "[config] please init config")
	return cfg
}

func DefaultFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Destination: &CfgPath,
			Aliases:     typex.StrOf("c"),
			Usage:       "config path",
			Value:       CfgPath,
		},
		&cli.StringFlag{
			Name:        "addr",
			Destination: &runenv.Addr,
			Aliases:     typex.StrOf("a"),
			Usage:       "server(http|grpc|ws|...) address",
			Value:       runenv.Addr,
		},
		&cli.BoolFlag{
			Name:        "trace",
			Destination: &runenv.Trace,
			Aliases:     typex.StrOf("t"),
			Usage:       "enable trace",
			Value:       runenv.Trace,
		},
		&cli.StringFlag{
			Name:        "mode",
			Destination: &runenv.Mode,
			Aliases:     typex.StrOf("m"),
			Usage:       "running mode(dev|test|stag|prod|release)",
			Value:       runenv.Mode,
		},
		&cli.StringFlag{
			Name:        "level",
			Destination: &runenv.Level,
			Aliases:     typex.StrOf("l"),
			Usage:       "log level(debug|info|warn|error|panic|fatal)",
			Value:       runenv.Level,
		},
		&cli.BoolFlag{
			Name:        "catch-sigpipe",
			Destination: &runenv.CatchSigpipe,
			Usage:       "catch and ignore SIGPIPE on stdout and stderr if specified",
			Value:       runenv.CatchSigpipe,
		},
	}
}
