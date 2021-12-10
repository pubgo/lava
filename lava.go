package lava

import (
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/internal/runtime"
	"reflect"
	"sync"
)

func Run(description string, entries ...entry.Entry) {
	runtime.Run(description, entries...)
}

func init() {
	sync.Pool{}
	var dd = reflect.TypeOf()
}
