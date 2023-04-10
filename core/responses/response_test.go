package responses

import (
	"encoding/json"
	"testing"

	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/proto/errorpb"
	"github.com/pubgo/funk/proto/testcodepb"
	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	err := errutil.ParseError(testcodepb.ErrCodeDbConn)
	rsp := &Response{Code: errorpb.Code_Internal, Message: "internal error", Detail: err, Data: nil}
	bytes, err1 := json.Marshal(rsp)
	assert.Nil(t, err1)
	t.Log(string(bytes))
}
