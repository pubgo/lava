package bootstrap

import (
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/orm"
	"github.com/pubgo/lava/servers/https"
)

type Config struct {
	Http   *https.Config   `yaml:"http"`
	Db     *orm.Config     `yaml:"db"`
	Metric *metrics.Config `yaml:"metric"`
	Log    *logging.Config `yaml:"logger"`
}
