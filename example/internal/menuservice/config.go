package menuservice

import (
	"errors"
	"fmt"
)

type defaultMethod struct {
	Path   string `mapstructure:"path"`
	Method string `mapstructure:"method"`
}

type Config struct {
	Path         string           `mapstructure:"path"`
	PrintRoute   bool             `mapstructure:"print-route"`
	DefaultMenus []*defaultMethod `mapstructure:"default-menus"`
}

func (c Config) Check() error {
	if c.Path == "" {
		return errors.New("menu path is null")
	}

	for _, m := range c.DefaultMenus {
		if m.Method == "" || m.Path == "" {
			return fmt.Errorf("method and path should not be null, data=>%v", m)
		}
	}
	return nil
}
