package rest

import (
	fb "github.com/pubgo/lug/builder/fiber"

	"github.com/pubgo/xlog"
)

const Name = "rest_entry"

var logs = xlog.GetLogger(Name)

type Cfg struct {
	fb.Cfg
}
