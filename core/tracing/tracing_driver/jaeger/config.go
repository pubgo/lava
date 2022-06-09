package jaeger

import (
	"fmt"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/xerror"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type Cfg struct {
	*jaegerCfg.Configuration
	Logger    *lumberjack.Logger
	BatchSize int32 `yaml:"batch_size"`
}

const Name = "jaeger"

func DefaultCfg() Cfg {
	cfg, err := jaegerCfg.FromEnv()
	xerror.Exit(err)

	cfg.Disabled = false
	cfg.ServiceName = runmode.Project
	xerror.Panic(err)
	return Cfg{
		Configuration: cfg,
		BatchSize:     1,
		Logger: &lumberjack.Logger{
			Filename:   fmt.Sprintf("./logs/trace/%s.log", runmode.Project),
			MaxSize:    50, // mb
			MaxBackups: 10,
			MaxAge:     1, //days
		},
	}
}
