package jaeger

import (
	"fmt"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/runmode"

	jaegerCfg "github.com/uber/jaeger-client-go/config"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type Cfg struct {
	*jaegerCfg.Configuration
	Logger    *lumberjack.Logger `yaml:"logger"`
	BatchSize int32              `yaml:"batch_size"`
}

const Name = "jaeger"

func DefaultCfg() Cfg {
	cfg, err := jaegerCfg.FromEnv()
	assert.Exit(err)

	cfg.Disabled = false
	cfg.ServiceName = runmode.Project
	assert.Must(err)
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
