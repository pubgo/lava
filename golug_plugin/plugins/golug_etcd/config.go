package golug_etcd

import (
	"time"
)

type Cfg struct {
	Configs map[string]ClientCfg `yaml:"configs" json:"configs" toml:"configs"`
}

type ClientCfg struct {
	Enabled              bool          `json:"enabled" yaml:"enabled"`
	Endpoints            []string      `json:"endpoints" yaml:"endpoints"`
	AutoSyncInterval     time.Duration `json:"auto_sync_interval" yaml:"auto_sync_interval"`
	DialTimeout          time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	DialKeepAliveTime    time.Duration `json:"dial_keep_alive_time" yaml:"dial_keep_alive_time"`
	DialKeepAliveTimeout time.Duration `json:"dial_keep_alive_timeout" yaml:"dial_keep_alive_timeout"`
	Username             string        `json:"username" yaml:"username"`
	Password             string        `json:"password" yaml:"password"`
	RejectOldCluster     bool          `json:"reject_old_cluster" yaml:"reject_old_cluster"`
	PermitWithoutStream  bool          `json:"permit_without_stream" yaml:"permit_without_stream"`
}
