package config

import (
	"testing"

	"github.com/pubgo/funk/pretty"
	"github.com/stretchr/testify/assert"
)

func TestConfigPath(t *testing.T) {
	t.Log(getConfigPath("", ""))
	assert.Panics(t, func() {
		t.Log(getConfigPath("", "toml"))
	})
}

func TestNew(t *testing.T) {
	assert.NotNil(t, New())
}

var _ NamedConfig = (*configL)(nil)

type configL struct {
	Name  string
	Value string
}

func (c configL) ConfigUniqueName() string {
	return c.Name
}

type configA struct {
	Names []*configL
	Name1 configL
}

func TestMerge(t *testing.T) {
	cfg := &configA{}
	assert.Nil(t, Merge(
		&configA{},
		configA{
			Name1: configL{
				Name: "a1",
			},
		},
	))
	t.Logf("%#v", cfg)

	cfg := &configA{}
	assert.Nil(t, Merge(
		cfg,
		configA{
			Name1: configL{
				Name: "a1",
			},
		},
		configA{
			Names: []*configL{
				{Name: "a2"},
			},
			Name1: configL{
				Name: "a2",
			},
		},
	))
	t.Logf("%#v", cfg)

	cfg = new(configA)
	assert.Nil(t, Merge(
		cfg,
		configA{
			Name1: configL{
				Name: "a1",
			},
		},

		configA{
			Names: []*configL{
				{Name: "a2", Value: "a2"},
			},
			Name1: configL{
				Name: "a2",
			},
		},

		configA{
			Names: []*configL{
				{Name: "a2", Value: "a3"},
				{Name: "a3"},
			},
			Name1: configL{
				Name: "a3",
			},
		},
	))
	pretty.Println(cfg)
	t.Logf("%#v", cfg)
}
