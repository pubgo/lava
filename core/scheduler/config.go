package scheduler

import (
	"time"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/v2/result"
	"github.com/reugn/go-quartz/quartz"
	"github.com/samber/lo"
)

func createConfig(configs []*Config) (r result.Result[map[string]*JobConfig]) {
	configMap := make(map[string]*JobConfig)
	if len(configs) == 0 || configs[0] == nil {
		return r.WithValue(configMap)
	}

	for _, config := range configs[0].JobConfigs {
		if config.Name == "" {
			return r.WithErrorf("schedule job name is empty")
		}

		if _, ok := configMap[config.Name]; ok {
			return r.WithErrorf("schedule job(%s) exists", config.Name)
		}
	}
	return r.WithValue(configMap)
}

func initConfig(name string, cfg *JobConfig, mergeCfg *JobConfig) (r result.Result[*JobConfig]) {
	if cfg == nil {
		cfg = &JobConfig{Name: name}
	}

	if cfg.Disabled == nil {
		cfg.Disabled = lo.ToPtr(false)
	}

	if cfg.Timeout == nil {
		cfg.Timeout = lo.ToPtr(time.Second * 10)
	}

	if cfg.RetryInterval == nil {
		cfg.RetryInterval = lo.ToPtr(time.Second)
	}

	if cfg.MaxRetries == nil {
		cfg.MaxRetries = lo.ToPtr(0)
	}

	if cfg.Replace == nil {
		cfg.Replace = lo.ToPtr(false)
	}

	if cfg.Location == nil {
		cfg.Location = lo.ToPtr(time.UTC.String())
	}
	cfg.location = result.Wrap(time.LoadLocation(lo.FromPtr(cfg.Location))).
		InspectErr(func(err error) {
			log.Err(err).Msgf("failed to parse time location:%s", lo.FromPtr(cfg.Location))
		}).
		UnwrapErr(&r)
	if r.IsErr() {
		return
	}

	if mergeCfg != nil {
		if mergeCfg.Disabled != nil {
			cfg.Disabled = mergeCfg.Disabled
		}

		if mergeCfg.Timeout != nil {
			cfg.Timeout = mergeCfg.Timeout
		}

		if mergeCfg.RetryInterval != nil {
			cfg.RetryInterval = mergeCfg.RetryInterval
		}

		if mergeCfg.MaxRetries != nil {
			cfg.MaxRetries = mergeCfg.MaxRetries
		}

		if mergeCfg.Replace != nil {
			cfg.Replace = mergeCfg.Replace
		}

		if mergeCfg.Location != nil {
			cfg.Location = mergeCfg.Location
		}
		cfg.location = result.Wrap(time.LoadLocation(lo.FromPtr(cfg.Location))).
			InspectErr(func(err error) {
				log.Err(err).Msgf("failed to parse time location:%s", lo.FromPtr(cfg.Location))
			}).
			UnwrapErr(&r)
		if r.IsErr() {
			return
		}
	}

	return r.WithValue(cfg)
}

type JobConfig struct {
	Disabled *bool          `yaml:"disabled"`
	Name     string         `yaml:"name"`
	Timeout  *time.Duration `yaml:"timeout"`

	// MaxRetries is the maximum number of retries before aborting the
	// current job execution.
	// Default: 0.
	MaxRetries *int `yaml:"max_retries"`

	// RetryInterval is the fixed time interval between retry attempts.
	// Default: 1 second.
	RetryInterval *time.Duration `yaml:"retry_interval"`

	// Replace indicates whether the job should replace an existing job
	// with the same key.
	// Default: false.
	Replace *bool `yaml:"replace"`

	Location *string `yaml:"location"`

	location *time.Location
}

func (c JobConfig) ToJobDetailOptions() *quartz.JobDetailOptions {
	return &quartz.JobDetailOptions{
		MaxRetries:    lo.FromPtr(c.MaxRetries),
		RetryInterval: lo.FromPtr(c.RetryInterval),
		Replace:       lo.FromPtr(c.Replace),
		Suspended:     false,
	}
}

type Config struct {
	Timeout    string      `yaml:"timeout"`
	JobConfigs []JobConfig `yaml:"jobs"`
}

type JobsConfigLoader struct {
	Scheduler *Config `yaml:"scheduler"`
}
