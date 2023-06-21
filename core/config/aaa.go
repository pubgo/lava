package config

import (
	"gopkg.in/yaml.v3"
)

const (
	defaultConfigName = "config"
	defaultConfigType = "yaml"
	defaultConfigPath = "./configs"
)

var (
	configDir  string
	configPath string
)

var _ yaml.Unmarshaler = (*Node)(nil)

type Node struct {
	maps  map[string]any
	value *yaml.Node
}

func (c *Node) UnmarshalYAML(value *yaml.Node) error {
	if c.maps == nil {
		c.maps = make(map[string]any)
	}

	if err := value.Decode(&c.maps); err != nil {
		return err
	}

	c.value = value

	return nil
}

func (c *Node) IsNil() bool {
	return c.value == nil
}

func (c *Node) Get(key string) any   { return c.maps[key] }
func (c *Node) Decode(val any) error { return c.value.Decode(val) }

type NamedConfig interface {
	// ConfigUniqueName unique name
	ConfigUniqueName() string
}

type Resources struct {
	Resources      []string `yaml:"resources"`
	PatchResources []string `yaml:"patch_resources"`
}
