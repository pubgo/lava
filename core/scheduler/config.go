package scheduler

import (
	"fmt"
	"time"

	"github.com/pubgo/funk/v2/result"
	"github.com/samber/lo"
)

func createConfig(opts []*Config) (r result.Result[map[string]*JobSetting]) {
	configMap := make(map[string]*JobSetting)
	if len(opts) > 0 && opts[0] != nil {
		for _, setting := range opts[0].JobSettings {
			if setting.Name == "" {
				return r.WithErrorf("schedule job name is empty")
			}

			if _, ok := configMap[setting.Name]; ok {
				return r.WithErrorf("schedule job(%s) exists", setting.Name)
			}

			configMap[setting.Name] = initConfig(setting.Name, lo.ToPtr(setting)).
				MapErr(func(err error) error {
					return fmt.Errorf("schedule job(%s) error: %w", setting.Name, err)
				}).
				UnwrapErr(&r)
			if r.IsErr() {
				return
			}
		}
	}
	return r.WithValue(configMap)
}

func initConfig(name string, cfg *JobSetting) (r result.Result[*JobSetting]) {
	if cfg == nil {
		cfg = &JobSetting{Name: name}
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
		cfg.location = result.Wrap(time.LoadLocation(lo.FromPtr(cfg.Location))).UnwrapErr(&r)
	}

	return r.WithValue(cfg)
}

type JobSetting struct {
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

	//quartz.JobDetailOptions
}

type Config struct {
	Timeout     string       `yaml:"timeout"`
	JobSettings []JobSetting `yaml:"jobs"`
}

type JobsConfigLoader struct {
	Scheduler *Config `yaml:"scheduler"`
}

type JobMetadata struct {
	Name          string
	Timeout       time.Duration
	MaxRetries    int
	RetryInterval time.Duration
	Replace       bool
	Location      *time.Location
}
