package bootstrap

import (
	"github.com/pubgo/funk/clients/orm"
	"github.com/pubgo/funk/metric"
	"github.com/pubgo/lava/logging/logconfig"
)

type Config struct {
	Db     *orm.Cfg          `yaml:"orm"`
	Metric *metric.Cfg       `yaml:"metric"`
	Log    *logconfig.Config `yaml:"logger"`
}
