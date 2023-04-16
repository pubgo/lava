package responses

import (
	"encoding/json"
	"testing"

	"github.com/pubgo/funk/proto/errorpb"
	"github.com/pubgo/funk/proto/testcodepb"
	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	rsp := &Response{
		Err: &errorpb.Error{
			Code: testcodepb.ErrCodeDbConn,
			Msg: &errorpb.ErrMsg{
				Msg:    "internal error",
				Detail: nil,
			}},
		Data: nil}
	bytes, err1 := json.Marshal(rsp)
	assert.Nil(t, err1)
	t.Log(string(bytes))
}
