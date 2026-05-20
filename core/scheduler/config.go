package scheduler

import (
	"time"

	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/result"
	"github.com/reugn/go-quartz/quartz"
	"github.com/samber/lo"
)

func createConfig(configs []*Config) (map[string]*JobConfig, error) {
	configMap := make(map[string]*JobConfig)
	if len(configs) == 0 || configs[0] == nil {
		return configMap, nil
	}

	for i := range configs[0].JobConfigs {
		config := &configs[0].JobConfigs[i]
		if config.Name == "" {
			return nil, errors.Errorf("schedule job name is empty")
		}

		if _, ok := configMap[config.Name]; ok {
			return nil, errors.Errorf("schedule job(%s) exists", config.Name)
		}
		configMap[config.Name] = config
	}
	return configMap, nil
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

func initAndMergeConfig(name string, jobConfigs ...*JobConfig) *JobConfig {
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

		if cfg.Location != nil {
			location, err := result.WrapErr(time.LoadLocation(lo.FromPtr(cfg.Location)))
			err.MustWithLog(func(e result.Event) { e.Msgf("failed to parse time location:%s", lo.FromPtr(cfg.Location)) })
			cfg.location = location
		}
	}

	return cfg
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
