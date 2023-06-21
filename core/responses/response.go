package responses

import (
	"github.com/pubgo/funk/proto/errorpb"
)

type Response struct {
	Err  *errorpb.ErrCode `json:"err"`
	Data any              `json:"data"`
}
