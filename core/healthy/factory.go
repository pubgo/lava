package healthy

import (
	"github.com/pubgo/funk/assert"

	"github.com/pubgo/lava/internal/pkg/typex"
	"github.com/pubgo/lava/internal/pkg/utils"
)

const Name = "health"

var healthList typex.SMap

func Get(names ...string) Handler {
	val, ok := healthList.Load(utils.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(Handler)
}

func List() (val []Handler) {
	healthList.Range(func(_, value interface{}) bool {
		val = append(val, value.(Handler))
		return true
	})
	return
}

func Register(name string, r Handler) {
	assert.If(name == "" || r == nil, "[name,r] is null")
	assert.If(healthList.Has(name), "healthy [%s] already exists", name)
	healthList.Set(name, r)
}
