package casbinservice

import (
	"errors"
)

type Config struct {
	Model          string `mapstructure:"model"`
	EnableLog      bool   `mapstructure:"enable-log"`
	EnableAutoSave bool   `mapstructure:"enable-auto-save"`
}

func (c Config) Check() error {
	if c.Model == "" {
		return errors.New("casbin model file not found")
	}

	return nil
}
