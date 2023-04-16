package responses

import (
	"github.com/pubgo/funk/proto/errorpb"
)

type Response struct {
	Err  *errorpb.Error `json:"err"`
	Data any            `json:"data"`
}
