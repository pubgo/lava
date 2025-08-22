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

func defaultConfig(name string) *JobConfig {
	return &JobConfig{
		Name:          name,
		Disabled:      lo.ToPtr(false),
		Timeout:       lo.ToPtr(time.Second * 10),
		RetryInterval: lo.ToPtr(time.Second),
		MaxRetries:    lo.ToPtr(3),
		Replace:       lo.ToPtr(false),
		Location:      lo.ToPtr(time.UTC.String()),
		location:      time.UTC,
	}
}

func initAndMergeConfig(name string, jobConfigs ...*JobConfig) (r result.Result[*JobConfig]) {
	cfg := defaultConfig(name)
	for _, jobConfig := range jobConfigs {
		if jobConfig == nil {
			continue
		}

		if jobConfig.Disabled != nil {
			cfg.Disabled = jobConfig.Disabled
		}

		if jobConfig.Timeout != nil {
			cfg.Timeout = jobConfig.Timeout
		}

		if jobConfig.RetryInterval != nil {
			cfg.RetryInterval = jobConfig.RetryInterval
		}

		if jobConfig.MaxRetries != nil {
			cfg.MaxRetries = jobConfig.MaxRetries
		}

		if jobConfig.Replace != nil {
			cfg.Replace = jobConfig.Replace
		}

		if jobConfig.Location != nil {
			cfg.Location = jobConfig.Location
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
