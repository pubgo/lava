package bootstrap

import (
	"github.com/pubgo/funk/clients/orm"
	"github.com/pubgo/funk/metric"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/servers/grpcs"
)

type Config struct {
	Grpc   *grpcs.Config   `yaml:"grpc"`
	Db     *orm.Cfg        `yaml:"orm"`
	Metric *metric.Cfg     `yaml:"metric"`
	Log    *logging.Config `yaml:"logger"`
}
