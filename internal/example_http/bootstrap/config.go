package bootstrap

import (
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/core/orm"
	"github.com/pubgo/lava/servers/grpcs"
)

type Config struct {
	Grpc   *grpcs.Config   `yaml:"grpc"`
	Db     *orm.Config     `yaml:"orm"`
	Metric *metric.Config  `yaml:"metric"`
	Log    *logging.Config `yaml:"logger"`
}
