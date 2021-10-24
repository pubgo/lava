package nacos

import (
	"time"
)

const Name = "nacos"

var cfgMap = make(map[string]*Cfg)

type Cfg struct {
	ClientConfig  *ClientConfig  `json:"client"`
	ServerConfigs []ServerConfig `json:"servers"`
	c             *Client
}

type ServerConfig struct {
	Scheme      string `json:"scheme"`
	ContextPath string `json:"context_path"`
	IpAddr      string `json:"ip_addr"`
	Port        uint64 `json:"port"`
}

type ClientConfig struct {
	TimeoutMs            uint64                   `json:"timeout_ms"`
	ListenInterval       uint64                   `json:"listen_interval"`
	BeatInterval         int64                    `json:"beat_interval"`
	NamespaceId          string                   `json:"namespace_id"`
	AppName              string                   `json:"app_name"`
	Endpoint             string                   `json:"endpoint"`
	RegionId             string                   `json:"region_id"`
	AccessKey            string                   `json:"access_key"`
	SecretKey            string                   `json:"secret_key"`
	OpenKMS              bool                     `json:"open_kms"`
	CacheDir             string                   `json:"cache_dir"`
	UpdateThreadNum      int                      `json:"update_thread_num"`
	NotLoadCacheAtStart  bool                     `json:"not_load_cache_at_start"`
	UpdateCacheWhenEmpty bool                     `json:"update_cache_when_empty"`
	Username             string                   `json:"username"`
	Password             string                   `json:"password"`
	LogDir               string                   `json:"log_dir"`
	RotateTime           string                   `json:"rotate_time"`
	MaxAge               int64                    `json:"max_age"`
	LogLevel             string                   `json:"log_level"`
	LogSampling          *ClientLogSamplingConfig `json:"log_sampling"`
	ContextPath          string                   `json:"context_path"`
}

type ClientLogSamplingConfig struct {
	Initial    int           `json:"initial"`
	Thereafter int           `json:"thereafter"`
	Tick       time.Duration `json:"tick"`
}
