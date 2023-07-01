package bootstrap

import (
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/servers/grpcs"

	"github.com/pubgo/lava/internal/example/grpc/services/gid_client"
)

type Config struct {
	grpcs.GrpcServerConfigLoader `yaml:",inline"`
	metrics.MetricConfigLoader   `yaml:",inline"`
	logging.LogConfigLoader      `yaml:",inline"`

	GidCli  *gid_client.GrpcConfig `yaml:"gid-client"`
	GidCli1 *gid_client.HttpConfig `yaml:"gid-client-http"`
}
