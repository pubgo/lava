package errutil

import (
	"encoding/json"

	"github.com/pubgo/lava/pkg/typex"
)

type Err struct {
	Msg    string  `json:"msg"`
	Detail typex.M `json:"detail"`
}

func (e *Err) Error() string {
	var dt, err = json.Marshal(e)
	if err != nil {
		return err.Error()
	}
	return string(dt)
}
