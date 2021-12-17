package restc

import (
	"github.com/pubgo/lava/types"
)

var _ types.Response = (*response)(nil)

type response struct {
	resp *Response
}

func (r *response) Header() types.Header { return types.Header(r.resp.Header) }
func (r *response) Payload() interface{} { return nil }
func (r *response) Stream() bool         { return false }
