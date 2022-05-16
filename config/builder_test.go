package config

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"

	"github.com/pubgo/lava/pkg/env"
)

func TestName(t *testing.T) {
	var assert = assertions.New(t)

	env.Set("env_prefix", "hello")
	env.Set("hello_123", "app.name=hello")

	var c = newCfg()
	assert.So(c, should.NotBeNil)

	assert.So(c.GetString("app.name"), should.Equal, "hello")
	assert.So(c.GetString("app.home"), should.Equal, c.GetString("app.project"))
}
