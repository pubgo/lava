package config

import (
	"github.com/pubgo/funk/pretty"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	cfg := Merge(
		configA{},
		configA{
			Name1: configL{
				Name: "a1",
			},
		},
	).Unwrap()
	t.Logf("%#v", cfg)

	cfg = Merge(
		configA{},
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
	).Unwrap()
	t.Logf("%#v", cfg)

	cfg = Merge(
		configA{},
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
	).Unwrap()
	pretty.Println(cfg)
	t.Logf("%#v", cfg)
}
