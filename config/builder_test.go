package config

import (
	"os"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"

	"github.com/pubgo/lava/runtime"
)

func TestName(t *testing.T) {
	var assert = assertions.New(t)

	var c = New()
	assert.So(c, should.NotBeNil)

	_ = os.Setenv(runtime.Project+"_123", "app.name=hello")
	c = New()
	assert.So(c.GetString("app.name"), should.Equal, "hello")
	assert.So(c.GetString("app.home"), should.Equal, c.GetString("app.project"))
}
