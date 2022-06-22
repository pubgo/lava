package config

import (
	"testing"

	"github.com/pubgo/lava/internal/pkg/env"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	var is = assert.New(t)

	env.Set("lava_hello", "hello")

	var c = newCfg()
	is.NotNil(c)
	is.Equal(c.GetString("hello"), "hello")
}
