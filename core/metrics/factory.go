package metrics

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
)

var factories = make(map[string]Factory)

func Get(name string) Factory  { return factories[name] }
func List() map[string]Factory { return factories }
func Register(name string, driver Factory) {
	defer recovery.Exit()
	assert.If(name == "" || driver == nil, "[driver,name] should not be null")
	assert.If(factories[name] != nil, "[driver] %s already exists", name)
	factories[name] = driver
}
