package quartz

import (
	"github.com/pubgo/x/async"
	_ "github.com/reugn/go-quartz/quartz"
)

func init() {
	async.Go()
}