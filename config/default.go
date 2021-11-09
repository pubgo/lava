package config

import (
	"path/filepath"
	"strings"

	"github.com/pubgo/xerror"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/types"
)

var (
	CfgType = "yaml"
	CfgName = "config"
	Home    = filepath.Join(xerror.PanicStr(filepath.Abs(filepath.Dir(""))), ".lava")
	CfgPath = ""
	cfg     = &configImpl{v: viper.New()}
)

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
			EnvVars:     types.EnvOf("trace", "trace-log", "tracelog"),
		},
		// 运行环境
		&cli.StringFlag{
			Name:        "mode",
			Destination: &runenv.Mode,
			Aliases:     typex.StrOf("m"),
			Usage:       "running mode(dev|test|stag|prod|release)",
			Value:       runenv.Mode,
			EnvVars:     types.EnvOf("run-mode", "run-env"),
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
