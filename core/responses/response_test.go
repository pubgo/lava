package responses

import (
	"encoding/json"
	"testing"

	"github.com/pubgo/funk/proto/testcodepb"
	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	rsp := &Response{
		Err:  testcodepb.ErrCodeDbConn,
		Data: nil,
	}
	bytes, err1 := json.Marshal(rsp)
	assert.Nil(t, err1)
	t.Log(string(bytes))
}
