package bootstrap

import (
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/core/orm"
	"github.com/pubgo/lava/servers/grpcs"
)

type Config struct {
	Grpc   *grpcs.Config   `yaml:"grpc"`
	Db     *orm.Config     `yaml:"orm"`
	Metric *metrics.Config `yaml:"metric"`
	Log    *logging.Config `yaml:"logger"`
}
