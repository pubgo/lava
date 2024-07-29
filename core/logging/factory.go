package logging

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
)

type Factory func(log log.Logger)

var factories = make(map[string]Factory)

func List() map[string]Factory { return factories }
func Register(name string, factory Factory) {
	defer recovery.Exit()
	assert.If(name == "" || factory == nil, "[factory, name] should not be null")
	assert.If(factories[name] != nil, "[factory] %s already exists", name)
	factories[name] = factory
}
