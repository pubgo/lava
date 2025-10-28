package healthy

import (
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/typex"
)

const Name = "healthy"

var healthList typex.SyncMap

func Get(name string) Handler {
	val, ok := healthList.Load(name)
	if !ok {
		return nil
	}

	return val.(Handler)
}

func List() (names []string) {
	healthList.Range(func(name, _ any) bool {
		names = append(names, name.(string))
		return true
	})
	return names
}

func Register(name string, r Handler) {
	assert.If(name == "" || r == nil, "[name,r] is null")
	assert.If(healthList.Has(name), "healthy [%s] already exists", name)
	healthList.Set(name, r)
}
