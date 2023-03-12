package responses

import (
	"github.com/pubgo/funk/proto/errorpb"
)

type Response struct {
	Code    errorpb.Code   `json:"code"`
	Message string         `json:"message"`
	Detail  *errorpb.Error `json:"detail"`
	Data    any            `json:"data"`
}
