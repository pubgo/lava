package bootstrap

import (
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/internal/example/grpc/services/gidclient"
	"github.com/pubgo/lava/servers/grpcs"
)

type Config struct {
	GrpcServer *grpcs.Config     `yaml:"grpc_server"`
	Metric     *metric.Config    `yaml:"metric"`
	Log        *logging.Config   `yaml:"logger"`
	GidCli     *gidclient.Config `yaml:"gid-client"`
}
